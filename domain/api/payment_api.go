package api

import "time"

type Payment struct {
	AccountID    string    `json:"account_id"`
	DueDate      time.Time `json:"due_date"`
	FullPayment  int       `json:"full_payment"`
	PaymentTitle string    `json:"payment_title"`
	Remark       *string   `json:"remark"`
	StartDate    time.Time `json:"start_date"`
}

type Transaction struct {
	AccountID     string `json:"account_id"`
	PaymentAmount int    `json:"payment_amount"`
}

type UpdatePaymentStatus struct {
	PaymentID int `json:"payment_id"`
}
