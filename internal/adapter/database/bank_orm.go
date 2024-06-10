package database

import (
	"time"

	"github.com/google/uuid"
)

type BankAccountOrm struct {
	AccountUuid    uuid.UUID `gorm:"primary_key"`
	AccountNumber  string
	AccountName    string
	Currency       string
	CurrentBalance float64
	Transactions   []BankTransactionOrm `gorm:"foreignKey:AccountUuid"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (BankAccountOrm) TableName() string {
	return "bank_accounts"
}

type BankTransactionOrm struct {
	TransactionUuid      uuid.UUID `gorm:"primary_key"`
	AccountUuid          uuid.UUID
	TransactionTimestamp time.Time
	Amount               float64
	TransactionType      string
	Notes                string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (BankTransactionOrm) TableName() string {
	return "bank_transactions"
}

type BankExchangeRateOrm struct {
	ExchangeRateUuid   uuid.UUID `gorm:"primary_key"`
	FromCurrency       string
	ToCurrency         string
	Rate               float64
	ValidFromTimestamp time.Time
	ValidToTimestamp   time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (BankExchangeRateOrm) TableName() string {
	return "bank_exchange_rates"
}

type BankTransferOrm struct {
	TransferUuid      uuid.UUID `gorm:"primary_key"`
	FromAccountUuid   uuid.UUID
	ToAccountUuid     uuid.UUID
	Currency          string
	Amount            float64
	TransferTimestamp time.Time
	TransferSuccess   bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (BankTransferOrm) TableName() string {
	return "bank_transfers"
}
