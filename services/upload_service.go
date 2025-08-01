package services

import (
	"database/sql"
	"fmt"
	"mime/multipart"
	"nhj-poc/database"
	"nhj-poc/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func ProcessExcelUpload(c *gin.Context) ([]models.Account, []models.Customer, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, nil, fmt.Errorf("no file uploaded: %w", err)
	}

	openedFile, err := file.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer openedFile.Close()

	return parseAndInsertExcel(openedFile)
}

func parseAndInsertExcel(file multipart.File) ([]models.Account, []models.Customer, error) {
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read Excel file: %w", err)
	}

	sheet := xlsx.GetSheetName(0)
	rows, err := xlsx.GetRows(sheet)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read sheet rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("no data found in Excel file")
	}

	headerMap := make(map[string]int)
	for i, col := range rows[0] {
		headerMap[col] = i
	}

	var Accounts []models.Account
	var Customers []models.Customer

	// Fetch all occupations from the database and populate the map
	occupations := []models.Occupation{}
	if err := database.DB.Model(&models.Occupation{}).Find(&occupations).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to fetch occupations: %w", err)
	}

	// Initialize the occupationsMap
	occupationsMap := make(map[string]*int)
	for _, occupation := range occupations {
		occupationsMap[occupation.OccupationName] = &occupation.OccupationID
	}

	for _, row := range rows[1:] {
		// Prepare the Account entry
		accountEntry := models.Account{
			AccountID:         mapRowToValue(row, headerMap, "accountId"),
			CustomerID:        StringPtrToNullString(mapRowToNullableValue(row, headerMap, "customerId")),
			ProductType:       StringPtrToNullString(mapRowToNullableValue(row, headerMap, "productType")),
			OutstandingAmount: convertIntPtrToNullInt64(mapRowToNullableInt(row, headerMap, "outstandingAmount")),
			OverdueAmount:     convertIntPtrToNullInt64(mapRowToNullableInt(row, headerMap, "overDueAmount")),
			DaysPastDue:       convertIntPtrToNullInt64(mapRowToNullableInt(row, headerMap, "daysPastDue")),
			SelfCured:         StringPtrToNullString(mapRowToNullableValue(row, headerMap, "selfCured")),
			TopUpScore:        StringPtrToNullString(mapRowToNullableValue(row, headerMap, "topUpScore")),
			LossOnSale:        convertIntPtrToNullInt64(mapRowToNullableInt(row, headerMap, "lossOnSale")),
			LossOnClaim:       StringPtrToNullString(mapRowToNullableValue(row, headerMap, "lossOnClaim")),
		}
		Accounts = append(Accounts, accountEntry)

		occupationName := mapRowToNullableValue(row, headerMap, "occupation")
		var occupationID *int
		if occupationName != nil {
			occupationID = occupationsMap[*occupationName]
		}

		// Prepare the Customer entry
		customerEntry := models.Customer{
			CustomerID:         mapRowToValue(row, headerMap, "customerId"),
			CustomerName:       StringPtrToNullString(mapRowToNullableValue(row, headerMap, "customerName")),
			OccupationID:       convertIntPtrToNullInt64(occupationID),
			RegisterAddress:    StringPtrToNullString(mapRowToNullableValue(row, headerMap, "registerAddress")),
			RegisterTambol:     StringPtrToNullString(mapRowToNullableValue(row, headerMap, "registerTambol")),
			RegisterAmphur:     StringPtrToNullString(mapRowToNullableValue(row, headerMap, "registerAmphur")),
			RegisterProvince:   StringPtrToNullString(mapRowToNullableValue(row, headerMap, "registerProvince")),
			RegisterPostalCode: convertIntPtrToNullInt64(mapRowToNullableInt(row, headerMap, "registerPostalCode")),
			CurrentAddress:     StringPtrToNullString(mapRowToNullableValue(row, headerMap, "currentAddress")),
			CurrentTambol:      StringPtrToNullString(mapRowToNullableValue(row, headerMap, "currentTambol")),
			CurrentAmphur:      StringPtrToNullString(mapRowToNullableValue(row, headerMap, "currentAmphur")),
			CurrentProvince:    StringPtrToNullString(mapRowToNullableValue(row, headerMap, "currentProvince")),
			CurrentPostalCode:  convertIntPtrToNullInt64(mapRowToNullableInt(row, headerMap, "currentPostalCode")),
		}

		Customers = append(Customers, customerEntry)
	}

	// Remove duplicates from Customers slice by CustomerID
	Customers = removeDuplicateCustomers(Customers)

	// Select existing customer IDs to avoid duplicates
	var existingCustomerIDs []string
	if err := database.DB.Model(&models.Customer{}).
		Select("customer_id").Where("customer_id IN (?)", getCustomerIDs(Customers)).Pluck("customer_id", &existingCustomerIDs).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to select existing customer IDs: %w", err)
	}

	// Update or insert customers based on existence
	var updatedCustomers []models.Customer
	var newCustomers []models.Customer
	for _, customer := range Customers {
		if contains(existingCustomerIDs, customer.CustomerID) {
			if err := database.DB.Model(&models.Customer{}).Where("customer_id = ?", customer.CustomerID).Updates(customer).Error; err != nil {
				return nil, nil, fmt.Errorf("failed to update customer with ID %s: %w", customer.CustomerID, err)
			}
			updatedCustomers = append(updatedCustomers, customer)
		} else {
			newCustomers = append(newCustomers, customer)
		}
	}

	// Select existing account IDs to avoid duplicates
	var existingAccountIDs []string
	if err := database.DB.Model(&models.Account{}).
		Select("account_id").Where("account_id IN (?)", getAccountIDs(Accounts)).Pluck("account_id", &existingAccountIDs).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to select existing account IDs: %w", err)
	}

	// Update or insert accounts based on existence
	var updatedAccounts []models.Account
	var newAccounts []models.Account
	for _, account := range Accounts {
		if contains(existingAccountIDs, account.AccountID) {
			if err := database.DB.Model(&models.Account{}).Where("account_id = ?", account.AccountID).Updates(account).Error; err != nil {
				return nil, nil, fmt.Errorf("failed to update account with ID %s: %w", account.AccountID, err)
			}
			updatedAccounts = append(updatedAccounts, account)
		} else {
			newAccounts = append(newAccounts, account)
		}
	}

	// Perform batch insert for new customer records
	if len(newCustomers) > 0 {
		if err := database.DB.CreateInBatches(newCustomers, 100).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to insert new customer batch: %w", err)
		}
	}

	// Perform batch insert for new account records
	if len(newAccounts) > 0 {
		if err := database.DB.CreateInBatches(newAccounts, 100).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to insert new account batch: %w", err)
		}
	}

	return newAccounts, newCustomers, nil
}

func mapRowToValue(row []string, headerMap map[string]int, field string) string {
	get := func(col string) string {
		i, ok := headerMap[col]
		if !ok || i >= len(row) {
			return ""
		}
		return row[i]
	}

	return get(field)
}

func mapRowToNullableValue(row []string, headerMap map[string]int, field string) *string {
	get := func(col string) string {
		i, ok := headerMap[col]
		if !ok || i >= len(row) {
			return ""
		}
		return row[i]
	}

	value := get(field)
	if value == "" {
		return nil
	}

	return &value
}

func mapRowToNullableInt(row []string, headerMap map[string]int, field string) *int {
	get := func(col string) string {
		i, ok := headerMap[col]
		if !ok || i >= len(row) {
			return ""
		}
		return row[i]
	}

	value := get(field)
	if value == "" {
		return nil
	}

	value = strings.ReplaceAll(value, ",", "")

	result, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &result
}

func removeDuplicateCustomers(Customers []models.Customer) []models.Customer {
	seen := make(map[string]models.Customer)
	result := []models.Customer{}

	for _, customerEntry := range Customers {
		seen[customerEntry.CustomerID] = customerEntry
	}

	for _, customerEntry := range seen {
		result = append(result, customerEntry)
	}

	return result
}

func getCustomerIDs(Customers []models.Customer) []string {
	var ids []string
	for _, customer := range Customers {
		ids = append(ids, customer.CustomerID)
	}
	return ids
}

func getAccountIDs(Accounts []models.Account) []string {
	var ids []string
	for _, account := range Accounts {
		ids = append(ids, account.AccountID)
	}
	return ids
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func convertIntPtrToNullInt64(i *int) *sql.NullInt64 {
	if i == nil {
		return &sql.NullInt64{Valid: false}
	}
	return &sql.NullInt64{Int64: int64(*i), Valid: true}
}

func StringPtrToNullString(s *string) *sql.NullString {
	ns := &sql.NullString{}
	if s != nil {
		ns.String = *s
		ns.Valid = true
	} else {
		ns.Valid = false
	}
	return ns
}

func compareAndUpdateAccount(account models.Account) error {
	var existingAccount models.Account
	if err := database.DB.Where("account_id = ?", account.AccountID).First(&existingAccount).Error; err != nil {
		return nil // Account does not exist, so we will insert it
	}

	// Compare each field and update only if necessary
	if account.ProductType != existingAccount.ProductType ||
		account.OutstandingAmount != existingAccount.OutstandingAmount ||
		account.OverdueAmount != existingAccount.OverdueAmount ||
		account.DaysPastDue != existingAccount.DaysPastDue ||
		account.SelfCured != existingAccount.SelfCured ||
		account.TopUpScore != existingAccount.TopUpScore ||
		account.LossOnSale != existingAccount.LossOnSale ||
		account.LossOnClaim != existingAccount.LossOnClaim {

		if err := database.DB.Model(&existingAccount).Updates(account).Error; err != nil {
			return fmt.Errorf("failed to update account with ID %s: %w", account.AccountID, err)
		}
	}
	return nil
}

func compareAndUpdateCustomer(customer models.Customer) error {
	var existingCustomer models.Customer
	if err := database.DB.Where("customer_id = ?", customer.CustomerID).First(&existingCustomer).Error; err != nil {
		return nil // Customer does not exist, so we will insert it
	}

	// Compare each field and update only if necessary
	if customer.CustomerName != existingCustomer.CustomerName ||
		customer.OccupationID != existingCustomer.OccupationID ||
		customer.RegisterAddress != existingCustomer.RegisterAddress ||
		customer.RegisterTambol != existingCustomer.RegisterTambol ||
		customer.RegisterAmphur != existingCustomer.RegisterAmphur ||
		customer.RegisterProvince != existingCustomer.RegisterProvince ||
		customer.RegisterPostalCode != existingCustomer.RegisterPostalCode ||
		customer.CurrentAddress != existingCustomer.CurrentAddress ||
		customer.CurrentTambol != existingCustomer.CurrentTambol ||
		customer.CurrentAmphur != existingCustomer.CurrentAmphur ||
		customer.CurrentProvince != existingCustomer.CurrentProvince ||
		customer.CurrentPostalCode != existingCustomer.CurrentPostalCode {

		if err := database.DB.Model(&existingCustomer).Updates(customer).Error; err != nil {
			return fmt.Errorf("failed to update customer with ID %s: %w", customer.CustomerID, err)
		}
	}
	return nil
}
