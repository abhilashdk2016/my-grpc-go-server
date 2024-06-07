package application

import (
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
