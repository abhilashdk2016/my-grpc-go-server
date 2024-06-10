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
	CreateTransfer(transfer database.BankTransferOrm) (uuid.UUID, error)
	CreateTransferTransactionPair(fromAccountOrm database.BankAccountOrm, toAccountOrm database.BankAccountOrm, fromTransactionOrm database.BankTransactionOrm, toTransactionOrm database.BankTransactionOrm) (bool, error)
	UpdateBankTransferStatus(transfer database.BankTransferOrm, status bool) error
}
