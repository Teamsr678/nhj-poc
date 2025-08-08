package service

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"nhj-poc/database"
	"nhj-poc/domain/entity"
	"nhj-poc/domain/model"
	"nhj-poc/util"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func ProcessExcelUpload(c *gin.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return fmt.Errorf("no file uploaded: %w", err)
	}

	openedFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer openedFile.Close()

	return parseAndInsertExcel(openedFile)
}

func parseAndInsertExcel(file multipart.File) error {
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read Excel file: %w", err)
	}

	sheet := xlsx.GetSheetName(0)
	rows, err := xlsx.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("cannot read sheet rows: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("no data found in Excel file")
	}

	headerMap := make(map[string]int)
	for i, col := range rows[0] {
		headerMap[col] = i
	}

	var Accounts []model.Account
	var Customers []model.Customer

	// Fetch all occupations from the database and populate the map
	occupations := []entity.Occupation{}
	if err := database.DB.Model(&entity.Occupation{}).Find(&occupations).Error; err != nil {
		return fmt.Errorf("failed to fetch occupations: %w", err)
	}

	occupationsMap := make(map[string]*int32)
	for _, occupation := range occupations {
		occupationsMap[occupation.OccupationName] = &occupation.OccupationID
	}

	for _, row := range rows[1:] {
		accountEntry := model.Account{
			AccountID:         mapRowToValue(row, headerMap, "accountId"),
			CustomerID:        mapRowToValue(row, headerMap, "customerId"),
			ProductType:       mapRowToNullableValue(row, headerMap, "productType"),
			OutstandingAmount: mapRowToNullableInt(row, headerMap, "outstandingAmount"),
			OverdueAmount:     mapRowToNullableInt(row, headerMap, "overDueAmount"),
			DaysPastDue:       mapRowToNullableInt(row, headerMap, "daysPastDue"),
			SelfCured:         mapRowToNullableValue(row, headerMap, "selfCured"),
			TopUpScore:        mapRowToNullableValue(row, headerMap, "topUpScore"),
			LossOnSale:        mapRowToNullableInt(row, headerMap, "lossOnSale"),
			LossOnClaim:       mapRowToNullableValue(row, headerMap, "lossOnClaim"),
			EarlyOA:           mapRowToNullableValue(row, headerMap, "earlyOA"),
		}
		Accounts = append(Accounts, accountEntry)

		occupationName := mapRowToNullableValue(row, headerMap, "occupation")
		var occupationID *int32
		if occupationName != nil {
			occupationID = occupationsMap[*occupationName]
		}

		customerEntry := model.Customer{
			CustomerID:         mapRowToValue(row, headerMap, "customerId"),
			CustomerName:       mapRowToNullableValue(row, headerMap, "customerName"),
			OccupationID:       occupationID,
			RegisterAddress:    mapRowToNullableValue(row, headerMap, "registerAddress"),
			RegisterTambol:     mapRowToNullableValue(row, headerMap, "registerTambol"),
			RegisterAmphur:     mapRowToNullableValue(row, headerMap, "registerAmphur"),
			RegisterProvince:   mapRowToNullableValue(row, headerMap, "registerProvince"),
			RegisterPostalCode: mapRowToNullableValue(row, headerMap, "registerPostalCode"),
			CurrentAddress:     mapRowToNullableValue(row, headerMap, "currentAddress"),
			CurrentTambol:      mapRowToNullableValue(row, headerMap, "currentTambol"),
			CurrentAmphur:      mapRowToNullableValue(row, headerMap, "currentAmphur"),
			CurrentProvince:    mapRowToNullableValue(row, headerMap, "currentProvince"),
			CurrentPostalCode:  mapRowToNullableValue(row, headerMap, "currentPostalCode"),
		}

		Customers = append(Customers, customerEntry)

	}

	// Remove duplicates from Customers slice by CustomerID
	Customers = removeDuplicateCustomers(Customers)

	// Select existing customer IDs to avoid duplicates
	var existingCustomers []entity.Customer
	if err := database.DB.Model(&entity.Customer{}).
		Where("customer_id IN (?)", getCustomerIDs(Customers)).
		Find(&existingCustomers).Error; err != nil {
		return fmt.Errorf("failed to select existing customer IDs: %w", err)
	}

	// Create a map for quick lookup of existing customers by customer_id
	existingCustomerMap := make(map[string]entity.Customer)
	for _, customer := range existingCustomers {
		existingCustomerMap[customer.CustomerID] = customer
	}

	// Update or insert customers based on existence
	var newCustomers []entity.Customer
	for _, customer := range Customers {
		existingCustomer, exists := existingCustomerMap[customer.CustomerID]

		if exists {
			if err := compareAndUpdateCustomer(existingCustomer, customer); err != nil {
				return fmt.Errorf("failed to compare and update customer: %w", err)
			}
		} else {
			c := ConvertModelToEntityCustomer(customer)
			newCustomers = append(newCustomers, c)
		}
	}

	// Perform batch insert for new customer records
	if len(newCustomers) > 0 {
		if err := database.DB.CreateInBatches(newCustomers, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert new customer batch: %w", err)
		}
	}

	// Select existing accounts to avoid duplicates
	var existingAccounts []entity.Account
	if err := database.DB.Model(&entity.Account{}).
		Where("account_id IN (?)", getAccountIDs(Accounts)).
		Find(&existingAccounts).Error; err != nil {
		return fmt.Errorf("failed to fetch existing account data: %w", err)
	}

	// Create a map for quick lookup of existing accounts by account_id
	existingAccountMap := make(map[string]entity.Account)
	for _, account := range existingAccounts {
		existingAccountMap[account.AccountID] = account
	}

	// Update or insert accounts based on existence
	var newAccounts []entity.Account
	for _, account := range Accounts {
		existingAccount, exists := existingAccountMap[account.AccountID]

		if exists {
			if err := compareAndUpdateAccount(existingAccount, account); err != nil {
				return fmt.Errorf("failed to compare and update account: %w", err)
			}
		} else {
			ac := ConvertModelToEntityAccount(account)
			newAccounts = append(newAccounts, ac)
		}
	}

	// Perform batch insert for new account records
	if len(newAccounts) > 0 {
		if err := database.DB.CreateInBatches(newAccounts, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert new account batch: %w", err)
		}
	}

	return nil
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

func mapRowToNullableInt(row []string, headerMap map[string]int, field string) *int32 {
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
	r := int32(result)

	return &r
}

func removeDuplicateCustomers(Customers []model.Customer) []model.Customer {
	seen := make(map[string]model.Customer)
	result := []model.Customer{}

	for _, customerEntry := range Customers {
		seen[customerEntry.CustomerID] = customerEntry
	}

	for _, customerEntry := range seen {
		result = append(result, customerEntry)
	}

	return result
}

func getCustomerIDs(Customers []model.Customer) []string {
	var ids []string
	for _, customer := range Customers {
		ids = append(ids, customer.CustomerID)
	}
	return ids
}

func getAccountIDs(Accounts []model.Account) []string {
	var ids []string
	for _, account := range Accounts {
		ids = append(ids, account.AccountID)
	}
	return ids
}

func compareAndUpdateAccount(existingAccount entity.Account, account model.Account) error {
	var updatedFields []string

	// mini-helper to fire off logDiff and track the field name
	recordChange := func(fieldName string, oldVal, newVal any) {
		logDiff(fieldName, oldVal, newVal)
		updatedFields = append(updatedFields, fieldName)
	}

	if !util.CompareNullable(existingAccount.CustomerID, account.CustomerID) {
		recordChange("CustomerID", existingAccount.CustomerID, account.CustomerID)
	}
	if !util.CompareNullable(existingAccount.ProductType, account.ProductType) {
		recordChange("ProductType", existingAccount.ProductType, account.ProductType)
	}
	if !util.CompareNullable(existingAccount.OutstandingAmount, account.OutstandingAmount) {
		recordChange("OutstandingAmount", existingAccount.OutstandingAmount, account.OutstandingAmount)
	}
	if !util.CompareNullable(existingAccount.OverdueAmount, account.OverdueAmount) {
		recordChange("OverdueAmount", existingAccount.OverdueAmount, account.OverdueAmount)
	}
	if !util.CompareNullable(existingAccount.DaysPastDue, account.DaysPastDue) {
		recordChange("DaysPastDue", existingAccount.DaysPastDue, account.DaysPastDue)
	}
	if !util.CompareNullable(existingAccount.SelfCured, account.SelfCured) {
		recordChange("SelfCured", existingAccount.SelfCured, account.SelfCured)
	}
	if !util.CompareNullable(existingAccount.TopUpScore, account.TopUpScore) {
		recordChange("TopUpScore", existingAccount.TopUpScore, account.TopUpScore)
	}
	if !util.CompareNullable(existingAccount.LossOnSale, account.LossOnSale) {
		recordChange("LossOnSale", existingAccount.LossOnSale, account.LossOnSale)
	}
	if !util.CompareNullable(existingAccount.LossOnClaim, account.LossOnClaim) {
		recordChange("LossOnClaim", existingAccount.LossOnClaim, account.LossOnClaim)
	}
	if !util.CompareNullable(existingAccount.EarlyOA, account.EarlyOA) {
		recordChange("EarlyOA", existingAccount.EarlyOA, account.EarlyOA)
	}

	if len(updatedFields) > 0 {
		log.Printf("Updating account %s; changed fields: %v",
			account.AccountID, updatedFields)
		if err := database.DB.
			Model(&existingAccount).
			Where("account_id = ?", account.AccountID).
			Updates(ConvertModelToEntityAccount(account)).
			Error; err != nil {
			return fmt.Errorf("failed to update account %s: %w",
				account.AccountID, err)
		}
	}

	return nil
}

func compareAndUpdateCustomer(existing entity.Customer, customer model.Customer) error {
	var updatedFields []string

	recordChange := func(fieldName string, oldVal, newVal any) {
		logDiff(fieldName, oldVal, newVal)
		updatedFields = append(updatedFields, fieldName)
	}

	if !util.CompareNullable(existing.CustomerName, customer.CustomerName) {
		recordChange("CustomerName", existing.CustomerName, customer.CustomerName)
	}
	if !util.CompareNullable(existing.OccupationID, customer.OccupationID) {
		recordChange("OccupationID", existing.OccupationID, customer.OccupationID)
	}
	if !util.CompareNullable(existing.RegisterAddress, customer.RegisterAddress) {
		recordChange("RegisterAddress", existing.RegisterAddress, customer.RegisterAddress)
	}
	if !util.CompareNullable(existing.RegisterTambol, customer.RegisterTambol) {
		recordChange("RegisterTambol", existing.RegisterTambol, customer.RegisterTambol)
	}
	if !util.CompareNullable(existing.RegisterAmphur, customer.RegisterAmphur) {
		recordChange("RegisterAmphur", existing.RegisterAmphur, customer.RegisterAmphur)
	}
	if !util.CompareNullable(existing.RegisterProvince, customer.RegisterProvince) {
		recordChange("RegisterProvince", existing.RegisterProvince, customer.RegisterProvince)
	}
	if !util.CompareNullable(existing.RegisterPostalCode, customer.RegisterPostalCode) {
		recordChange("RegisterPostalCode", existing.RegisterPostalCode, customer.RegisterPostalCode)
	}
	if !util.CompareNullable(existing.CurrentAddress, customer.CurrentAddress) {
		recordChange("CurrentAddress", existing.CurrentAddress, customer.CurrentAddress)
	}
	if !util.CompareNullable(existing.CurrentTambol, customer.CurrentTambol) {
		recordChange("CurrentTambol", existing.CurrentTambol, customer.CurrentTambol)
	}
	if !util.CompareNullable(existing.CurrentAmphur, customer.CurrentAmphur) {
		recordChange("CurrentAmphur", existing.CurrentAmphur, customer.CurrentAmphur)
	}
	if !util.CompareNullable(existing.CurrentProvince, customer.CurrentProvince) {
		recordChange("CurrentProvince", existing.CurrentProvince, customer.CurrentProvince)
	}
	if !util.CompareNullable(existing.CurrentPostalCode, customer.CurrentPostalCode) {
		recordChange("CurrentPostalCode", existing.CurrentPostalCode, customer.CurrentPostalCode)
	}

	if len(updatedFields) > 0 {
		log.Printf("Updating customer %s; changed fields: %v",
			customer.CustomerID, updatedFields)
		if err := database.DB.
			Model(&existing).
			Where("customer_id = ?", customer.CustomerID).
			Updates(ConvertModelToEntityCustomer(customer)).
			Error; err != nil {
			return fmt.Errorf("failed to update customer %s: %w",
				customer.CustomerID, err)
		}
	}

	return nil
}

func logDiff(fieldName string, oldVal any, newVal any) {
	var oldStr string
	switch v := oldVal.(type) {
	case *sql.NullString:
		if v != nil && v.Valid {
			oldStr = v.String
		}
	case *sql.NullInt32:
		if v != nil && v.Valid {
			oldStr = strconv.FormatInt(int64(v.Int32), 10)
		}
	default:
		oldStr = fmt.Sprint(v)
	}

	var newStr string
	switch v := newVal.(type) {
	case *string:
		if v != nil {
			newStr = *v
		}
	case *int32:
		if v != nil {
			newStr = strconv.FormatInt(int64(*v), 10)
		}
	default:
		newStr = fmt.Sprint(v)
	}

	log.Printf("%s changed: Old: %s, New: %s", fieldName, oldStr, newStr)
}

func ConvertModelToEntityCustomer(m model.Customer) entity.Customer {
	e := entity.Customer{
		CustomerID: m.CustomerID,
	}

	toNullString := func(s *string) *sql.NullString {
		if s != nil {
			return &sql.NullString{String: *s, Valid: true}
		}
		return &sql.NullString{Valid: false}
	}
	toNullInt32 := func(i *int32) *sql.NullInt32 {
		if i != nil {
			return &sql.NullInt32{Int32: *i, Valid: true}
		}
		return &sql.NullInt32{Valid: false}
	}

	e.CustomerName = toNullString(m.CustomerName)
	e.OccupationID = toNullInt32(m.OccupationID)
	e.RegisterAddress = toNullString(m.RegisterAddress)
	e.RegisterTambol = toNullString(m.RegisterTambol)
	e.RegisterAmphur = toNullString(m.RegisterAmphur)
	e.RegisterProvince = toNullString(m.RegisterProvince)
	e.RegisterPostalCode = toNullString(m.RegisterPostalCode)
	e.CurrentAddress = toNullString(m.CurrentAddress)
	e.CurrentTambol = toNullString(m.CurrentTambol)
	e.CurrentAmphur = toNullString(m.CurrentAmphur)
	e.CurrentProvince = toNullString(m.CurrentProvince)
	e.CurrentPostalCode = toNullString(m.CurrentPostalCode)

	return e
}

func ConvertModelToEntityAccount(m model.Account) entity.Account {
	e := entity.Account{
		AccountID: m.AccountID,
	}

	toNullString := func(s *string) *sql.NullString {
		if s != nil {
			return &sql.NullString{String: *s, Valid: true}
		}
		return &sql.NullString{Valid: false}
	}
	toNullInt32 := func(i *int32) *sql.NullInt32 {
		if i != nil {
			return &sql.NullInt32{Int32: *i, Valid: true}
		}
		return &sql.NullInt32{Valid: false}
	}

	e.CustomerID = m.CustomerID
	e.ProductType = toNullString(m.ProductType)
	e.OutstandingAmount = toNullInt32(m.OutstandingAmount)
	e.OverdueAmount = toNullInt32(m.OverdueAmount)
	e.DaysPastDue = toNullInt32(m.DaysPastDue)
	e.SelfCured = toNullString(m.SelfCured)
	e.TopUpScore = toNullString(m.TopUpScore)
	e.LossOnSale = toNullInt32(m.LossOnSale)
	e.LossOnClaim = toNullString(m.LossOnClaim)
	e.EarlyOA = toNullString(m.EarlyOA)

	return e
}
