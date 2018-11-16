package core

type Transaction struct {
	Name   string
	Amount float32
}

func NewTransaction(name string, amount float32) Transaction {
	return Transaction{
		Name:   name,
		Amount: amount}
}
