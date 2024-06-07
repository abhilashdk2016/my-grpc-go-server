package port

import (
	"time"

	"github.com/abhilashdk2016/my-grpc-go-server/internal/adapter/database"
	"github.com/google/uuid"
)

type DummyDatabasePort interface {
	Save(data *database.DummyOrm) (uuid.UUID, error)
	GetByUuid(uuid *uuid.UUID) (database.DummyOrm, error)
}

type BankDatabasePort interface {
	GetBankAccountByAccountNumber(acct string) (database.BankAccountOrm, error)
	CreateExchangeRate(r database.BankExchangeRateOrm) (uuid.UUID, error)
	GetExchangeRateAtTimestamp(fromCur string, toCur string, ts time.Time) (database.BankExchangeRateOrm, error)
	CreateTransaction(acct database.BankAccountOrm, t database.BankTransactionOrm) (uuid.UUID, error)
}
