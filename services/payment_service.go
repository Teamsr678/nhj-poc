package services

import (
	"fmt"
	"nhj-poc/constant"
	"nhj-poc/database"
	"nhj-poc/domain/entity"
	"nhj-poc/domain/model"
	"nhj-poc/util"

	"gorm.io/gorm"
)

func InsertPayment(paymentAPI model.Payment) error {
	exists, err := accountIDExists(database.DB, paymentAPI.AccountID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("payment for account_id %q not found", paymentAPI.AccountID)
	}

	paymentStatusID, err := getPaymentStatusID(database.DB, paymentAPI.AccountID)
	if err != nil {
		return err
	}

	if paymentStatusID != nil && *paymentStatusID == constant.Normal {
		return fmt.Errorf("payment for account_id %q is payment_status_id %q", paymentAPI.AccountID, constant.Normal)
	}

	paymentEntity := entity.Payment{
		AccountID:       paymentAPI.AccountID,
		DueDate:         paymentAPI.DueDate,
		FullPayment:     paymentAPI.FullPayment,
		PaymentStatusID: util.IntToNullInt32(constant.Normal),
	}
	if err := database.DB.Create(&paymentEntity).Error; err != nil {
		return err
	}

	return nil
}

func InsertTransaction() error {
	return nil
}

func UpdatePayment() error {
	return nil
}

func accountIDExists(db *gorm.DB, accountID string) (bool, error) {
	var count int64
	if err := db.
		Model(&entity.Account{}).
		Where("account_id = ?", accountID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func getPaymentStatusID(db *gorm.DB, accountID string) (*int32, error) {
	var payment entity.Payment
	if err := db.
		Model(&entity.Payment{}).
		Where("account_id = ?", accountID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if payment.PaymentStatusID == nil || !payment.PaymentStatusID.Valid {
		return nil, nil
	}

	return &payment.PaymentStatusID.Int32, nil
}

func getPaymentByPaymentID(db *gorm.DB, paymentID int) (*entity.Payment, error) {
	var payment entity.Payment
	if err := db.
		Model(&entity.Payment{}).
		Where("payment_id = ?", paymentID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	return &payment, nil
}

func GetTotalPaymentByPaymentID(db *gorm.DB, paymentID int) (*entity.TotalPaymentAmount, error) {
	var results entity.TotalPaymentAmount
	if err := db.
		Model(&entity.Transaction{}).
		Where("payment_id = ?", paymentID).
		Select("SUM(payment_amount) AS total_payment_amount").
		Group("payment_id").
		Scan(&results).Error; err != nil {
		return nil, err
	}
	return &results, nil
}

func UpdatePaymentStatusByID(paymentId int) error {
	payment, err := getPaymentByPaymentID(database.DB, paymentId)
	if err != nil {
		return err
	}
	if payment != nil && payment.PaymentStatusID.Int32 != constant.Normal {
		return fmt.Errorf("payment for payment_id %d is payment_status_id %d", paymentId, payment.PaymentStatusID.Int32)
	}

	totalPayment, err := GetTotalPaymentByPaymentID(database.DB, paymentId)
	if err != nil {
		return err
	}
	if totalPayment.TotalPaymentAmount == 0 {
		payment.PaymentStatusID = util.IntToNullInt32(constant.Broken)
	} else if payment.FullPayment <= totalPayment.TotalPaymentAmount {
		payment.PaymentStatusID = util.IntToNullInt32(constant.Full)
	} else {
		payment.PaymentStatusID = util.IntToNullInt32(constant.Partial)
	}
	if err := database.DB.
		Model(&entity.Payment{}).
		Where("payment_id = ?", paymentId).
		Updates(payment).
		Error; err != nil {
		return fmt.Errorf("failed to update payment %d: %w", paymentId, err)
	}
	return nil
}
