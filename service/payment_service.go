package service

import (
	"fmt"
	"nhj-poc/constant"
	"nhj-poc/database"
	"nhj-poc/domain/entity"
	"nhj-poc/domain/model"
	"nhj-poc/repository"
	"nhj-poc/util"
	"time"
)

func InsertPayment(pModel model.Payment) error {
	exists, err := repository.AccountIDExists(database.DB, pModel.AccountID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("account_id not found")
	}

	paymentEntity := entity.Payment{
		AccountID:       pModel.AccountID,
		DueDate:         pModel.DueDate,
		FullPayment:     pModel.FullPayment,
		PaymentStatusID: util.IntToNullInt32(constant.Normal),
		PaymentTitle:    pModel.PaymentTitle,
		Remark:          pModel.Remark,
		StartDate:       pModel.StartDate,
	}
	if err := database.DB.Create(&paymentEntity).Error; err != nil {
		return err
	}

	return nil
}

func InsertTransaction(tModel model.Transaction) error {
	exists, err := repository.AccountIDExists(database.DB, tModel.AccountID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("payment_id not found")
	}

	tEntity := entity.Transaction{
		AccountID:       tModel.AccountID,
		PaymentAmount:   tModel.PaymentAmount,
		TransactionDate: time.Now(),
	}

	if err := database.DB.Create(&tEntity).Error; err != nil {
		return err
	}

	return nil
}

func UpdatePaymentStatusByIDs(ids []int) error {
	var paymentIds []int
	var err error

	if len(ids) == 0 {
		paymentIds, err = repository.GetOverduePaymentIDs(database.DB)
		if err != nil {
			return fmt.Errorf("failed to get overdue payment IDs: %w", err)
		}
	}

	for _, paymentId := range paymentIds {
		if err := UpdatePaymentStatusByID(paymentId); err != nil {
			return fmt.Errorf("failed to update payment status for ID %d: %w", paymentId, err)
		}
	}

	return nil
}

func UpdatePaymentStatusByID(paymentId int) error {
	payment, err := repository.GetPaymentByPaymentID(database.DB, paymentId)
	if err != nil {
		return err
	}

	totalPayment, err := repository.GetTotalPayment(database.DB, payment.AccountID, payment.StartDate, payment.DueDate)
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
