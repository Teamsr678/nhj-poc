package services

import (
	"nhj-poc/database"
	"nhj-poc/domain/entity"
	"nhj-poc/domain/model"
	"nhj-poc/util"
)

func InsertPayment(paymentAPI model.Payment) error {
	paymentEntity := entity.Payment{
		PaymentID:       paymentAPI.PaymentID,
		AccountID:       paymentAPI.AccountID,
		DueDate:         paymentAPI.DueDate,
		FullPayment:     paymentAPI.FullPayment,
		PaymentStatusID: util.IntPtrToNullInt32(paymentAPI.PaymentStatusID),
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
