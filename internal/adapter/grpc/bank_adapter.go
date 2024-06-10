package grpc

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/abhilashdk2016/my-grpc-go-server/internal/application/domain/bank"
	bank_proto "github.com/abhilashdk2016/my-grpc-proto/protogen/go/bank-proto"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/datetime"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank_proto.CurrentBalanceRequest) (*bank_proto.CurrentBalanceResponse, error) {
	now := time.Now()
	bal := a.bankService.FindCurrentBalance(req.AccountNumber)
	return &bank_proto.CurrentBalanceResponse{
		Amount: bal,
		CurrentDate: &date.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}, nil
}

func (a *GrpcAdapter) FetchExchangeRates(req *bank_proto.ExchangeRateRequest, stream bank_proto.BankService_FetchExchangeRatesServer) error {
	context := stream.Context()

	for {
		select {
		case <-context.Done():
			log.Println("Client cancelled stream")
			return nil
		default:
			now := time.Now().Truncate(time.Second)
			rate := a.bankService.FindExchangeRate(req.FromCurrency, req.FromCurrency, now)

			stream.Send(
				&bank_proto.ExchangeRateResponse{
					FromCurrency: req.FromCurrency,
					ToCurrency:   req.ToCurrency,
					Rate:         rate,
					Timestamp:    now.Format(time.RFC3339),
				},
			)

			log.Printf("Exchange rate sent to client, %v to %v : %v\n", req.FromCurrency, req.ToCurrency, rate)
			time.Sleep(5 * time.Second)
		}

	}
}

func toTime(dt *datetime.DateTime) (time.Time, error) {
	if dt == nil {
		now := time.Now()

		dt = &datetime.DateTime{
			Year:    int32(now.Year()),
			Month:   int32(now.Month()),
			Day:     int32(now.Day()),
			Hours:   int32(now.Hour()),
			Minutes: int32(now.Minute()),
			Seconds: int32(now.Second()),
			Nanos:   int32(now.Nanosecond()),
		}
	}

	res := time.Date(int(dt.Year), time.Month(dt.Month), int(dt.Day),
		int(dt.Hours), int(dt.Minutes), int(dt.Seconds), int(dt.Nanos), time.UTC)

	return res, nil
}

func (a *GrpcAdapter) SummarizeTransactions(stream bank_proto.BankService_SummarizeTransactionsServer) error {
	tsum := bank.TransactionSummary{
		SummaryOnDate: time.Now(),
		SumIn:         0,
		SumOut:        0,
		SumTotal:      0,
	}

	acct := ""

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			res := bank_proto.TransactionSummary{
				AccountNumber: acct,
				SumAmountIn:   tsum.SumIn,
				SumAmountOut:  tsum.SumOut,
				SumTotal:      tsum.SumTotal,
				TransactionDate: &datetime.DateTime{
					Year:  int32(tsum.SummaryOnDate.Year()),
					Month: int32(tsum.SummaryOnDate.Month()),
					Day:   int32(tsum.SummaryOnDate.Day()),
				},
			}

			return stream.SendAndClose(&res)
		}

		if err != nil {
			log.Fatalln("Error while reading from client :", err)
		}

		acct = req.AccountNumber

		ts, err := toTime(req.Timestamp)

		if err != nil {
			log.Fatalf("Error while parsing timestamp %v : %v", req.Timestamp, err)
		}

		ttype := bank.TransactionTypeUnknown

		if req.Type == bank_proto.TransactionType_TRANSACION_TYPE_IN {
			ttype = bank.TransactionTypeIn
		} else if req.Type == bank_proto.TransactionType_TRANSACION_TYPE_OUT {
			ttype = bank.TransactionTypeOut
		}

		tcur := bank.Transaction{
			Amount:          req.Amount,
			Timestamp:       ts,
			TransactionType: ttype,
		}

		_, err = a.bankService.CreateTransaction(req.AccountNumber, tcur)

		if err != nil {
			log.Println("Error while creating transaction :", err)
		}

		err = a.bankService.CalculateTransactionSummary(&tsum, tcur)
		if err != nil {
			return err
		}
	}
}

func currentTime() *datetime.DateTime {
	now := time.Now()
	return &datetime.DateTime{
		Year:       int32(now.Year()),
		Month:      int32(now.Month()),
		Day:        int32(now.Day()),
		Hours:      int32(now.Hour()),
		Minutes:    int32(now.Minute()),
		Seconds:    int32(now.Second()),
		Nanos:      int32(now.Nanosecond()),
		TimeOffset: &datetime.DateTime_UtcOffset{},
	}
}

func (a *GrpcAdapter) TransferMultiple(stream bank_proto.BankService_TransferMultipleServer) error {
	context := stream.Context()
	for {
		select {
		case <-context.Done():
			log.Println("Client cancelled stream")
			return nil
		default:
			req, err := stream.Recv()

			if err == io.EOF {
				return nil
			}

			if err != nil {
				log.Fatalln("Error while reading from client :", err)
			}

			tt := bank.TrasferTransaction{
				FromAccountNumber: req.FromAccountNumber,
				ToAccountNumber:   req.ToAccountNumber,
				Currency:          req.Currency,
				Amount:            req.Amount,
			}

			_, transferSuccess, err := a.bankService.Transfer(tt)

			if err != nil {
				return err
			}

			res := bank_proto.TransferResponse{
				FromAccountNumber: req.FromAccountNumber,
				ToAccountNumber:   req.ToAccountNumber,
				Currency:          req.Currency,
				Amount:            req.Amount,
				Timestamp:         currentTime(),
			}

			if transferSuccess {
				res.Status = bank_proto.TransferStatus_TRANSFER_STATUS_SUCCESS
			} else {
				res.Status = bank_proto.TransferStatus_TRANSFER_STATUS_FAIL
			}

			err = stream.Send(&res)
			if err != nil {
				log.Fatalln("Error while sending response to client :", err)
			}
		}
	}
}
