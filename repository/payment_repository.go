package repository

import (
	"nhj-poc/constant"
	"nhj-poc/domain/entity"
	"time"

	"gorm.io/gorm"
)

func PaymentIDExists(db *gorm.DB, paymentID int) (bool, error) {
	var count int64
	if err := db.
		Model(&entity.Payment{}).
		Where("payment_id = ?", paymentID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetOverduePaymentIDs(db *gorm.DB) ([]int, error) {
	cutoff := time.Now().AddDate(0, 0, -3)

	var ids []int
	if err := db.
		Model(&entity.Payment{}).
		Where("due_date < ? AND payment_status_id = ?", cutoff, constant.Normal).
		Pluck("payment_id", &ids).Error; err != nil {
		return nil, err
	}

	return ids, nil
}

func GetPaymentStatusIDByPaymentID(db *gorm.DB, paymentID int) (*int32, error) {
	var payment entity.Payment
	if err := db.
		Model(&entity.Payment{}).
		Where("payment_id = ?", paymentID).
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

func GetPaymentByPaymentID(db *gorm.DB, paymentID int) (*entity.Payment, error) {
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

func AccountIDExists(db *gorm.DB, accountID string) (bool, error) {
	var count int64
	if err := db.
		Model(&entity.Account{}).
		Where("account_id = ?", accountID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetPaymentStatusID(db *gorm.DB, accountID string) (*int32, error) {
	var payment entity.Payment
	if err := db.
		Model(&entity.Payment{}).
		Where("account_id = ?", accountID).
		Last(&payment).Error; err != nil {
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

func DateIsOverlapping(db *gorm.DB, accountID string, startDate time.Time) (bool, error) {
	var count int64
	if err := db.
		Model(&entity.Payment{}).
		Where("account_id = ? AND ((start_date <= ? AND due_date >= ?))",
			accountID, startDate, startDate).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
