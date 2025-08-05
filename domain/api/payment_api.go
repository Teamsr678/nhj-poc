package api

import "time"

type Payment struct {
	PaymentID       int       `json:"payment_id"`
	AccountID       string    `json:"account_id"`
	DueDate         time.Time `json:"due_date"`
	FullPayment     string    `json:"full_payment"`
	PaymentStatusID *int      `json:"payment_status_id"`
}
