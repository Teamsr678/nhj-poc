package model

import "time"

type Payment struct {
	AccountID   string
	DueDate     time.Time
	FullPayment int
}

type Transaction struct {
	PaymentID     int
	PaymentAmount int
}
