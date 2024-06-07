package port

import (
	"time"

	dbank "github.com/abhilashdk2016/my-grpc-go-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

type BankServicePort interface {
	FindCurrentBalance(acct string) float64
	CreateExchangeRate(r dbank.ExchangeRate) (uuid.UUID, error)
	FindExchangeRate(fromCur string, toCur string, ts time.Time) float64
}
