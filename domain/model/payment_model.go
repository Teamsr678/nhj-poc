package model

import "time"

type Payment struct {
	PaymentID       int
	AccountID       string
	DueDate         time.Time
	FullPayment     string
	PaymentStatusID *int
}
