package port

type BankServicePort interface {
	FindCurrentBalance(acct string) float64
}
