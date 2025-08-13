package service

import (
	"database/sql"
	"fmt"
	"math"
	"nhj-poc/constant"
	"nhj-poc/database"
	"nhj-poc/domain/entity"
	"nhj-poc/repository"
)

type CapacityOA struct {
	OAId        string
	Capacity    int
	CapacityC2C int
	CapacityCRL int
}

func ToNullString(s *string) *sql.NullString {
	if s != nil && *s != "" {
		return &sql.NullString{String: *s, Valid: true}
	}
	return &sql.NullString{Valid: false}
}

func UpdateAssignmentsByProductType() error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Delete Assignments product type
	if err := repository.DeleteAssignments(tx, constant.ASSIGN_BY_PRODUCT_TYPE); err != nil {
		return fmt.Errorf("failed to delete assignments: %w", err)
	}

	//Get accounts data
	accounts, err := repository.GetAllAccount(tx)
	if err != nil {
		return fmt.Errorf("failed to get all accounts: %w", err)
	}
	var accountsBucket1 []entity.Account
	for _, account := range accounts {
		if account.DaysPastDue.Int32 >= constant.BUCKET_1_MIN_DPD && account.DaysPastDue.Int32 <= constant.BUCKET_1_MAX_DPD {
			accountsBucket1 = append(accountsBucket1, account)
		}
	}
	countC2C := 0
	countCRL := 0
	for _, account := range accountsBucket1 {
		if account.ProductType.String == constant.PRODUCT_TYPE_C2C {
			countC2C++
		} else if account.ProductType.String == constant.PRODUCT_TYPE_CRL {
			countCRL++
		}
	}

	//Get oa data
	oas, err := repository.GetAllOA(tx)
	if err != nil {
		return fmt.Errorf("failed to get all oa: %w", err)
	}

	var queueC2C []string
	var queueCRL []string
	capacityOA := make(map[string]CapacityOA)
	for _, oa := range oas {
		if oa.C2CPercentage.Float64 > 0 {
			queueC2C = append(queueC2C, oa.OAId)
		}
		if oa.CRLPercentage.Float64 > 0 {
			queueCRL = append(queueCRL, oa.OAId)
		}
		capacityOA[oa.OAId] = CapacityOA{
			OAId:        oa.OAId,
			Capacity:    int(oa.Capacity.Int16),
			CapacityC2C: int(math.Round(oa.C2CPercentage.Float64 * float64(countC2C))),
			CapacityCRL: int(math.Round(oa.CRLPercentage.Float64 * float64(countCRL))),
		}
	}

	var assignments []entity.Assignments
	productType := constant.ASSIGN_BY_PRODUCT_TYPE
	for _, account := range accountsBucket1 {
		var assignOaID string = ""
		if account.ProductType.String == constant.PRODUCT_TYPE_C2C {
			if len(queueC2C) > 0 {
				assignOaID = queueC2C[0]
				queueC2C = queueC2C[1:]
				capacityOA[assignOaID] = CapacityOA{
					OAId:        assignOaID,
					Capacity:    capacityOA[assignOaID].Capacity - 1,
					CapacityC2C: capacityOA[assignOaID].CapacityC2C - 1,
					CapacityCRL: capacityOA[assignOaID].CapacityCRL,
				}
				if capacityOA[assignOaID].Capacity > 0 {
					if capacityOA[assignOaID].CapacityC2C > 0 {
						queueC2C = append(queueC2C, assignOaID)
					}
				} else {
					var newQueueCRL []string
					for _, oa := range queueCRL {
						if oa != assignOaID {
							newQueueCRL = append(newQueueCRL, oa)
						}
					}
					queueCRL = newQueueCRL
				}
			}
		} else if account.ProductType.String == constant.PRODUCT_TYPE_CRL {
			if len(queueCRL) > 0 {
				assignOaID = queueCRL[0]
				queueCRL = queueCRL[1:]
				capacityOA[assignOaID] = CapacityOA{
					OAId:        assignOaID,
					Capacity:    capacityOA[assignOaID].Capacity - 1,
					CapacityC2C: capacityOA[assignOaID].CapacityC2C,
					CapacityCRL: capacityOA[assignOaID].CapacityCRL - 1,
				}
				if capacityOA[assignOaID].Capacity > 0 {
					if capacityOA[assignOaID].CapacityCRL > 0 {
						queueCRL = append(queueCRL, assignOaID)
					}
				} else {
					var newQueueC2C []string
					for _, oa := range queueC2C {
						if oa != assignOaID {
							newQueueC2C = append(newQueueC2C, oa)
						}
					}
					queueC2C = newQueueC2C
				}
			}
		}
		assignments = append(assignments, entity.Assignments{
			AccountID: ToNullString(&account.AccountID),
			OaID:      ToNullString(&assignOaID),
			AssignBy:  ToNullString(&productType),
		})
	}

	if err := tx.CreateInBatches(assignments, 1000).Error; err != nil {
		return fmt.Errorf("failed to insert new assignment batch: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
