package model

import "time"

type Payment struct {
	AccountID    string
	DueDate      time.Time
	FullPayment  int
	PaymentTitle string
	Remark       *string
	StartDate    time.Time
}

type Transaction struct {
	AccountID     string `json:"account_id"`
	PaymentAmount int
}
