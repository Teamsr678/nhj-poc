package entity

import (
	"database/sql"
	"time"
)

type Payment struct {
	PaymentID       int            `gorm:"column:payment_id;primaryKey;autoIncrement"`
	AccountID       string         `gorm:"column:account_id;type:text;not null"`
	DueDate         time.Time      `gorm:"column:due_date;type:date;not null"`
	FullPayment     int            `gorm:"column:full_payment;type:int;not null"`
	PaymentStatusID *sql.NullInt32 `gorm:"column:payment_status_id"`
}

func (Payment) TableName() string {
	return "payment"
}

type Transaction struct {
	TransactionID   int       `gorm:"column:transaction_id;primaryKey;autoIncrement"`
	PaymentID       int       `gorm:"column:payment_id;type:text;not null"`
	PaymentAmount   int       `gorm:"column:payment_amount;type:date;not null"`
	TransactionDate time.Time `gorm:"column:transaction_date;type:date;not null"`
}

func (Transaction) TableName() string {
	return "transaction"
}

type TotalPaymentAmount struct {
	TotalPaymentAmount int `gorm:"column:total_payment_amount"`
}
