package services

import (
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

	// Build header map
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
			CustomerID:        mapRowToValue(row, headerMap, "customerId"),
			ProductType:       mapRowToValue(row, headerMap, "productType"),
			OutstandingAmount: mapRowToNullableInt(row, headerMap, "outstandingAmount"),
			OverdueAmount:     mapRowToNullableInt(row, headerMap, "overDueAmount"),
			DaysPastDue:       mapRowToNullableInt(row, headerMap, "daysPastDue"),
			SelfCured:         mapRowToValue(row, headerMap, "selfCured"),
			TopUpScore:        mapRowToValue(row, headerMap, "topUpScore"),
			LossOnSale:        mapRowToNullableInt(row, headerMap, "lossOnSale"),
			LossOnClaim:       mapRowToValue(row, headerMap, "lossOnClaim"),
		}
		Accounts = append(Accounts, accountEntry)

		// Prepare the Customer entry
		customerEntry := models.Customer{
			CustomerID:         mapRowToValue(row, headerMap, "customerId"),
			CustomerName:       mapRowToValue(row, headerMap, "customerName"),
			OccupationID:       occupationsMap[mapRowToValue(row, headerMap, "occupation")],
			RegisterAddress:    mapRowToValue(row, headerMap, "registerAddress"),
			RegisterTambol:     mapRowToValue(row, headerMap, "registerTambol"),
			RegisterAmphur:     mapRowToValue(row, headerMap, "registerAmphur"),
			RegisterProvince:   mapRowToValue(row, headerMap, "registerProvince"),
			RegisterPostalCode: mapRowToNullableInt(row, headerMap, "registerPostalCode"),
			CurrentAddress:     mapRowToValue(row, headerMap, "currentAddress"),
			CurrentTambol:      mapRowToValue(row, headerMap, "currentTambol"),
			CurrentAmphur:      mapRowToValue(row, headerMap, "currentAmphur"),
			CurrentProvince:    mapRowToValue(row, headerMap, "currentProvince"),
			CurrentPostalCode:  mapRowToNullableInt(row, headerMap, "currentPostalCode"),
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

	// Remove duplicates from Customers slice by CustomerID
	var filteredCustomers []models.Customer
	for _, customer := range Customers {
		if !contains(existingCustomerIDs, customer.CustomerID) {
			filteredCustomers = append(filteredCustomers, customer)
		}
	}

	// Select existing account IDs to avoid duplicates
	var existingAccountIDs []string
	if err := database.DB.Model(&models.Account{}).
		Select("account_id").Where("account_id IN (?)", getAccountIDs(Accounts)).Pluck("account_id", &existingAccountIDs).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to select existing account IDs: %w", err)
	}

	// Remove duplicates from Accounts slice by AccountID
	var filteredAccounts []models.Account
	for _, account := range Accounts {
		if !contains(existingAccountIDs, account.AccountID) {
			filteredAccounts = append(filteredAccounts, account)
		}
	}

	// Perform batch insert for filtered customer records
	if len(filteredCustomers) > 0 {
		if err := database.DB.CreateInBatches(filteredCustomers, 100).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to insert customer_export batch: %w", err)
		}
	}

	// Perform batch insert for filtered account records
	if len(filteredAccounts) > 0 {
		if err := database.DB.CreateInBatches(filteredAccounts, 100).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to insert account_export batch: %w", err)
		}
	}

	return filteredAccounts, filteredCustomers, nil
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

func mapRowToNullableInt(row []string, headerMap map[string]int, field string) *int {
	get := func(col string) string {
		i, ok := headerMap[col]
		if !ok || i >= len(row) {
			return ""
		}
		return row[i]
	}

	// Fetch the value as a string and convert it
	value := get(field)
	if value == "" {
		return nil
	}

	// Remove commas from the value before converting it to an integer
	value = strings.ReplaceAll(value, ",", "")

	result, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &result
}

// removeDuplicateCustomers filters out customer export records with duplicate CustomerID.
func removeDuplicateCustomers(Customers []models.Customer) []models.Customer {
	seen := make(map[string]models.Customer) // To track unique CustomerIDs and store the latest record
	result := []models.Customer{}

	for _, customerEntry := range Customers {
		// Update the map with the latest record for this CustomerID
		seen[customerEntry.CustomerID] = customerEntry
	}

	// Now, append all unique Customer entries (which contain the last data for each CustomerID)
	for _, customerEntry := range seen {
		result = append(result, customerEntry)
	}

	return result
}

// getCustomerIDs returns a list of CustomerIDs from the Customers slice
func getCustomerIDs(Customers []models.Customer) []string {
	var ids []string
	for _, customer := range Customers {
		ids = append(ids, customer.CustomerID)
	}
	return ids
}

// getAccountIDs returns a list of AccountIDs from the Accounts slice
func getAccountIDs(Accounts []models.Account) []string {
	var ids []string
	for _, account := range Accounts {
		ids = append(ids, account.AccountID)
	}
	return ids
}

// contains checks if a string exists in the slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
