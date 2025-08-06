package api

import "time"

type Payment struct {
	AccountID   string    `json:"account_id"`
	DueDate     time.Time `json:"due_date"`
	FullPayment int       `json:"full_payment"`
}

type Transaction struct {
	PaymentID     int `json:"payment_id"`
	PaymentAmount int `json:"payment_amount"`
}

type UpdatePaymentStatus struct {
	PaymentID int `json:"payment_id"`
}
