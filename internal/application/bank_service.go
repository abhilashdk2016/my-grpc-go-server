package application

import (
	"fmt"
	"log"
	"time"

	"github.com/abhilashdk2016/my-grpc-go-server/internal/adapter/database"
	dbank "github.com/abhilashdk2016/my-grpc-go-server/internal/application/domain/bank"
	"github.com/abhilashdk2016/my-grpc-go-server/internal/port"
	"github.com/google/uuid"
)

type BankService struct {
	db port.BankDatabasePort
}

func NewBankService(dbPort port.BankDatabasePort) *BankService {
	return &BankService{
		db: dbPort,
	}
}

func (b *BankService) FindCurrentBalance(acct string) float64 {
	bankAccount, err := b.db.GetBankAccountByAccountNumber(acct)
	if err != nil {
		log.Println("Error in FindCurrentBalance :", err)
	}

	return bankAccount.CurrentBalance
}

func (b *BankService) CreateExchangeRate(r dbank.ExchangeRate) (uuid.UUID, error) {
	newUuid := uuid.New()
	now := time.Now()

	exchangeRateOrm := database.BankExchangeRateOrm{
		ExchangeRateUuid:   newUuid,
		FromCurrency:       r.FromCurrency,
		ToCurrency:         r.ToCurrency,
		Rate:               r.Rate,
		ValidFromTimestamp: r.ValidFromTimestamp,
		ValidToTimestamp:   r.ValidToTimestamp,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	return b.db.CreateExchangeRate(exchangeRateOrm)
}

func (b *BankService) FindExchangeRate(fromCur string, toCur string, ts time.Time) float64 {
	exchangeRate, err := b.db.GetExchangeRateAtTimestamp(fromCur, toCur, ts)

	if err != nil {
		return 0
	}

	return float64(exchangeRate.Rate)
}

func (b *BankService) CalculateTransactionSummary(tcur *dbank.TransactionSummary, trans dbank.Transaction) error {
	switch trans.TransactionType {
	case dbank.TransactionTypeIn:
		tcur.SumIn += trans.Amount
	case dbank.TransactionTypeOut:
		tcur.SumOut += trans.Amount
	default:
		return fmt.Errorf("unknown transaction type %v", trans.TransactionType)
	}

	tcur.SumTotal = tcur.SumIn - tcur.SumOut

	return nil
}

func (b *BankService) CreateTransaction(acct string, t dbank.Transaction) (uuid.UUID, error) {
	newuuid := uuid.New()
	now := time.Now()

	bankAccountOrm, err := b.db.GetBankAccountByAccountNumber(acct)

	if err != nil {
		log.Printf("Can't create transaction for %v : %v\n", acct, err)
		return uuid.Nil, err
	}

	transactionOrm := database.BankTransactionOrm{
		TransactionUuid:      newuuid,
		AccountUuid:          bankAccountOrm.AccountUuid,
		TransactionType:      t.TransactionType,
		TransactionTimestamp: now,
		Amount:               t.Amount,
		Notes:                t.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	savedUuid, err := b.db.CreateTransaction(bankAccountOrm, transactionOrm)

	return savedUuid, err
}

func (b *BankService) Transfer(tt dbank.TrasferTransaction) (uuid.UUID, bool, error) {
	now := time.Now()

	fromAccountOrm, err := b.db.GetBankAccountByAccountNumber(tt.FromAccountNumber)

	if err != nil {
		log.Printf("Can't find transfer from account %v : %v\n", tt.FromAccountNumber, err)
		return uuid.Nil, false, err
	}

	toAccountOrm, err := b.db.GetBankAccountByAccountNumber(tt.ToAccountNumber)

	if err != nil {
		log.Printf("Can't find transfer to account %v : %v\n", tt.ToAccountNumber, err)
		return uuid.Nil, false, err
	}

	fromTransactionOrm := database.BankTransactionOrm{
		TransactionUuid:      uuid.New(),
		TransactionTimestamp: now,
		TransactionType:      dbank.TransactionTypeOut,
		AccountUuid:          fromAccountOrm.AccountUuid,
		Amount:               tt.Amount,
		Notes:                "Transfer out to " + tt.ToAccountNumber,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	toTransactionOrm := database.BankTransactionOrm{
		TransactionUuid:      uuid.New(),
		TransactionTimestamp: now,
		TransactionType:      dbank.TransactionTypeIn,
		AccountUuid:          toAccountOrm.AccountUuid,
		Amount:               tt.Amount,
		Notes:                "Transfer in to " + tt.FromAccountNumber,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	newTransferUUid := uuid.New()

	transferOrm := database.BankTransferOrm{
		TransferUuid:      newTransferUUid,
		FromAccountUuid:   fromAccountOrm.AccountUuid,
		ToAccountUuid:     toAccountOrm.AccountUuid,
		Currency:          tt.Currency,
		Amount:            tt.Amount,
		TransferTimestamp: now,
		TransferSuccess:   false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if _, err := b.db.CreateTransfer(transferOrm); err != nil {
		log.Printf("Can't create transfer from %v to %v : %v\n", tt.FromAccountNumber, tt.ToAccountNumber, err)
		return uuid.Nil, false, err
	}

	if transferPairSuccess, err := b.db.CreateTransferTransactionPair(fromAccountOrm, toAccountOrm, fromTransactionOrm, toTransactionOrm); transferPairSuccess {
		b.db.UpdateBankTransferStatus(transferOrm, true)
		return newTransferUUid, true, nil
	} else {
		return newTransferUUid, false, err
	}
}
