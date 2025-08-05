package entity

import (
	"database/sql"
	"time"
)

type Payment struct {
	PaymentID       int
	AccountID       string
	DueDate         time.Time
	FullPayment     string
	PaymentStatusID *sql.NullInt32
}
