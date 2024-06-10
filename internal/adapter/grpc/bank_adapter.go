package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/abhilashdk2016/my-grpc-go-server/internal/application/domain/bank"
	bank_proto "github.com/abhilashdk2016/my-grpc-proto/protogen/go/bank-proto"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/datetime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank_proto.CurrentBalanceRequest) (*bank_proto.CurrentBalanceResponse, error) {
	now := time.Now()
	bal, err := a.bankService.FindCurrentBalance(req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"account %v not found", req.AccountNumber,
		)
	}
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
			rate, err := a.bankService.FindExchangeRate(req.FromCurrency, req.FromCurrency, now)

			if err != nil {
				s := status.New(codes.InvalidArgument, "Currency not valid. Please use valid currency for both to and from")
				s, _ = s.WithDetails(&errdetails.ErrorInfo{
					Domain: "my-grpc-bank-server",
					Reason: "INVALID_CURRENCY",
					Metadata: map[string]string{
						"from_currency": req.FromCurrency,
						"to_currency":   req.ToCurrency,
					},
				})

				return s.Err()
			}

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

		accountuuid, err := a.bankService.CreateTransaction(req.AccountNumber, tcur)

		if err != nil && accountuuid == uuid.Nil {
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "account_number",
						Description: "Invalid Account Number",
					},
				},
			})
			return s.Err()
		} else if err != nil && accountuuid != uuid.Nil {
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "amount",
						Description: fmt.Sprintf("Requested amount %v exceeds available banalnce", req.Amount),
					},
				},
			})
			return s.Err()
		}

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
				return buildTransferErrorStatusGrpc(err, *req)
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

func buildTransferErrorStatusGrpc(err error, req bank_proto.TransferRequest) error {
	switch {
	case errors.Is(err, bank.ErrTransferSourceAccountNotFound):
		s := status.New(codes.FailedPrecondition, err.Error())
		s, _ = s.WithDetails(&errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "INVALID_ACCOUNT",
					Subject:     "Source account not found",
					Description: fmt.Sprintf("source account (from %v) not found", req.FromAccountNumber),
				},
			},
		})

		return s.Err()
	case errors.Is(err, bank.ErrTransferDestinationAccountNotFound):
		s := status.New(codes.FailedPrecondition, err.Error())
		s, _ = s.WithDetails(&errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "INVALID_ACCOUNT",
					Subject:     "Destination account not found",
					Description: fmt.Sprintf("destination account (to %v) not found", req.ToAccountNumber),
				},
			},
		})

		return s.Err()
	case errors.Is(err, bank.ErrTransferRecordFailed):
		s := status.New(codes.Internal, err.Error())
		s, _ = s.WithDetails(&errdetails.Help{
			Links: []*errdetails.Help_Link{
				{
					Url:         "my-grpc-bank.com/faq",
					Description: "Bank FAQ",
				},
			},
		})

		return s.Err()
	case errors.Is(err, bank.ErrTransferTransactionPair):
		s := status.New(codes.InvalidArgument, err.Error())
		s, _ = s.WithDetails(&errdetails.ErrorInfo{
			Domain: "my-grpc-bank.com",
			Reason: "TRANSACTION_PAIR_FAILED",
			Metadata: map[string]string{
				"from_account": req.FromAccountNumber,
				"to_account":   req.ToAccountNumber,
				"currency":     req.Currency,
				"amount":       fmt.Sprintf("%f", req.Amount),
			},
		})

		return s.Err()
	default:
		s := status.New(codes.Unknown, err.Error())
		return s.Err()
	}
}
