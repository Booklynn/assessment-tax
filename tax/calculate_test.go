package tax

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestCalculateTaxValidRequest(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"donation"}).AddRow(100000)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"k-receipt"}).AddRow(50000)
	mock.ExpectQuery(`SELECT "k-receipt" FROM allowances WHERE id = ?`).WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 3000000,
		WHT:         0.0,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
			{AllowanceType: "k-receipt", Amount: 0},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody TaxPayable
	err := json.NewDecoder(rec.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)
	require.Equal(t, float64(639000), responseBody.Tax)
}

func TestCalculateTaxWithWTHReturnTaxRefund(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"donation"}).AddRow(100000)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 500000.1,
		WHT:         30000.1,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody TaxReturnable
	err := json.NewDecoder(rec.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)
	require.Equal(t, 1000.09, responseBody.TaxRefund)
}

func TestCalculateTaxRefundWhenHaveWTHButIncomeLessThanCriteria(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"donation"}).AddRow(100000)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 10,
		WHT:         30000.1,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 100000},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody TaxReturnable
	err := json.NewDecoder(rec.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)
	require.Equal(t, 30000.1, responseBody.TaxRefund)
}

func TestCalculateTaxWithWTHReturnTax(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"donation"}).AddRow(100000)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 500000.02,
		WHT:         25000.21,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody TaxPayable
	err := json.NewDecoder(rec.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)
	require.Equal(t, 3999.79, responseBody.Tax)
}

func TestCalculateTaxInvalidRequest(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	reqBodyJSON := `{"totalIncome": "not a number"}`

	req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(reqBodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NotEmpty(t, c)

	errorCalculateTax := CalculateTax(c)

	require.Error(t, errorCalculateTax)
}

func TestCalculateTaxWithNegativeTotalIncome(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: -1,
		WHT:         0.0,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "total income and wht cannot be less than 0", rec.Body.String())
}

func TestCalculateTaxWithInvalidAllowanceType(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 1,
		WHT:         1,
		Allowances: []Allowances{
			{AllowanceType: "kkkk", Amount: 0},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "allowanceType not allowed", rec.Body.String())
}

func TestCalculateTaxWithAllowanceTypeDuplication(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 1,
		WHT:         1,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
			{AllowanceType: "donation", Amount: 0},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "found allowanceType duplication", rec.Body.String())
}

func TestCalculateTaxWithErrorPersonalAllowance(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnError(sql.ErrNoRows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 1,
		WHT:         1,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
		},
	}

	_, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.Error(t, errorCalculateTax)
}

func TestCalculateTaxWithErrorDonationAllowance(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnError(sql.ErrNoRows)

	requestBody := TaxInfo{
		TotalIncome: 1,
		WHT:         1,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
		},
	}

	_, err := getAllowancesAmount(requestBody)

	require.Error(t, err)
}

func TestCalculateTaxWithErrorKReceiptAllowance(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	mock.ExpectQuery("SELECT k-receipt FROM allowances WHERE id = ?").WithArgs(1).WillReturnError(sql.ErrNoRows)

	requestBody := TaxInfo{
		TotalIncome: 1,
		WHT:         1,
		Allowances: []Allowances{
			{AllowanceType: "k-receipt", Amount: 0},
		},
	}

	_, err := getAllowancesAmount(requestBody)

	require.Error(t, err)
}

func TestCalculateTaxWithErrorGetAllowancesAmount(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 2000000,
		WHT:         0.0,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 200000},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCalculateTaxWithdonationAmountMoreThanSetting(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"donation"}).AddRow(100000)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"k-receipt"}).AddRow(50000)
	mock.ExpectQuery("SELECT k-receipt FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 2000000,
		WHT:         0.0,
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 200000},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody TaxPayable
	err := json.NewDecoder(rec.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)
	require.Equal(t, 278000.0, responseBody.Tax)
}

func TestCalculateTaxWithKReceiptAmountMoreThanSetting(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)
	rows = mock.NewRows([]string{"k-receipt"}).AddRow(50000)
	mock.ExpectQuery(`SELECT "k-receipt" FROM allowances WHERE id = ?`).WithArgs(1).WillReturnRows(rows)

	e := echo.New()
	requestBody := TaxInfo{
		TotalIncome: 1000000,
		WHT:         30000.1,
		Allowances: []Allowances{
			{AllowanceType: "k-receipt", Amount: 50001},
		},
	}

	rec, c := mockNewRequest(requestBody, t, e, "/tax/calculations")

	errorCalculateTax := CalculateTax(c)

	require.NoError(t, errorCalculateTax)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody TaxPayable
	err := json.NewDecoder(rec.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)
	require.Equal(t, 63499.9, responseBody.Tax)
}

func TestCalculateTaxByLevels(t *testing.T) {
	testCases := []struct {
		netIncome   float64
		expectedTax float64
	}{
		{netIncome: 0, expectedTax: 0},
		{netIncome: 1, expectedTax: 0},
		{netIncome: 1.1, expectedTax: 0},
		{netIncome: 150000, expectedTax: 0},
		{netIncome: 150000.02, expectedTax: 0},
		{netIncome: 150001, expectedTax: 0},
		{netIncome: 499999, expectedTax: 28999.9},
		{netIncome: 500000, expectedTax: 29000},
		{netIncome: 500001, expectedTax: 29000.1},
		{netIncome: 500001.51, expectedTax: 29000.15},
		{netIncome: 999999, expectedTax: 100999.85},
		{netIncome: 999999.99, expectedTax: 101000},
		{netIncome: 1000000, expectedTax: 101000},
		{netIncome: 1000001, expectedTax: 101000.15},
		{netIncome: 1999999, expectedTax: 297999.8},
		{netIncome: 2000000, expectedTax: 298000},
		{netIncome: 2000001, expectedTax: 298000.2},
		{netIncome: 2100000, expectedTax: 324000},
		{netIncome: 3000000, expectedTax: 639000},
		{netIncome: 3000001, expectedTax: 639000.35},
	}

	for _, tt := range testCases {
		actualTax := calculateTaxByLevels(tt.netIncome, 60000)

		require.Equal(t, tt.expectedTax, math.Round(actualTax*100)/100)
	}
}

func TestCheckTaxInfoNotNegative(t *testing.T) {
	var requestBody TaxInfo
	requestBody.TotalIncome = 0
	requestBody.WHT = 0

	err := checkTaxInfoNotNegative(requestBody)

	require.NoError(t, err)
}

func TestCheckTaxInfoNegativeReturnError(t *testing.T) {
	var requestBody TaxInfo
	requestBody.TotalIncome = -1500000
	requestBody.WHT = -1

	err := checkTaxInfoNotNegative(requestBody)

	require.Error(t, err)
	require.EqualError(t, err, "total income and wht cannot be less than 0")
}

func TestCheckValidTaxAllowanceType(t *testing.T) {
	taxInfo := TaxInfo{
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
			{AllowanceType: "k-receipt", Amount: 0},
		},
	}

	err := checkValidTaxAllowances(taxInfo)

	require.NoError(t, err)
}

func TestCheckInvalidAllowanceTypeReturnError(t *testing.T) {
	taxInfo := TaxInfo{
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: 0},
			{AllowanceType: "invalidType!", Amount: 0},
			{AllowanceType: "k-receipt", Amount: 0},
		},
	}

	err := checkValidTaxAllowances(taxInfo)

	require.Error(t, err)
	require.EqualError(t, err, "allowanceType not allowed")
}

func TestCheckTaxAllowanceAmountNegativeReturnError(t *testing.T) {
	taxInfo := TaxInfo{
		Allowances: []Allowances{
			{AllowanceType: "donation", Amount: -1},
		},
	}

	err := checkValidTaxAllowances(taxInfo)

	require.Error(t, err)
	require.EqualError(t, err, "allowance amount cannot be less than 0")
}

func mockNewRequest(requestBody TaxInfo, t *testing.T, e *echo.Echo, url string) (*httptest.ResponseRecorder, echo.Context) {
	reqBodyJSON, err := json.Marshal(requestBody)
	require.NoError(t, err)
	require.NotEmpty(t, reqBodyJSON)

	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(string(reqBodyJSON)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NotEmpty(t, c)
	return rec, c
}
