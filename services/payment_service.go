package services

import (
	"fmt"
	"nhj-poc/constant"
	"nhj-poc/database"
	"nhj-poc/domain/entity"
	"nhj-poc/domain/model"
	"nhj-poc/util"
	"time"

	"gorm.io/gorm"
)

func InsertPayment(pModel model.Payment) error {
	exists, err := accountIDExists(database.DB, pModel.AccountID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("account_id not found")
	}

	paymentStatusID, err := getPaymentStatusID(database.DB, pModel.AccountID)
	if err != nil {
		return err
	}

	if paymentStatusID != nil && *paymentStatusID == constant.Normal {
		return fmt.Errorf("Can't create payment because payment_status_id = '%d'", constant.Normal)
	}

	paymentEntity := entity.Payment{
		AccountID:       pModel.AccountID,
		DueDate:         pModel.DueDate,
		FullPayment:     pModel.FullPayment,
		PaymentStatusID: util.IntToNullInt32(constant.Normal),
	}
	if err := database.DB.Create(&paymentEntity).Error; err != nil {
		return err
	}

	return nil
}

func InsertTransaction(tModel model.Transaction) error {
	exists, err := paymentIDExists(database.DB, tModel.PaymentID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("payment_id not found")
	}

	paymentStatusID, err := getPaymentStatusIDByPaymentID(database.DB, tModel.PaymentID)
	if err != nil {
		return err
	}

	if paymentStatusID != nil && *paymentStatusID == constant.Normal {
		tEntity := entity.Transaction{
			PaymentID:       tModel.PaymentID,
			PaymentAmount:   tModel.PaymentAmount,
			TransactionDate: time.Now(),
		}

		if err := database.DB.Create(&tEntity).Error; err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Can't create transaction because payment_status_id='%d'", *paymentStatusID)
	}

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

func getPaymentStatusIDByPaymentID(db *gorm.DB, paymentID int) (*int32, error) {
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

func paymentIDExists(db *gorm.DB, paymentID int) (bool, error) {
	var count int64
	if err := db.
		Model(&entity.Payment{}).
		Where("payment_id = ?", paymentID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
