package tax

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSetPersonalAllowanceAmount(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	e := echo.New()
	requestBody := Allowances{
		Amount: 70000,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE allowances SET personal = $1 WHERE id = $2`)).
		WithArgs(requestBody.Amount, 1).WillReturnResult(sqlmock.NewResult(1, 1))

	rec, c := mockNewRequestAdmin(requestBody, t, e, "/admin/deductions/personal")

	err := SetPersonalAllowanceAmount(c)

	require.NoError(t, err)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSetPersonalAllowanceAmountWithInvalidRequest(t *testing.T) {
	db, _ := setupMockDB()
	conn = db

	e := echo.New()
	reqBodyJSON := `{"amount": "not a number"}`

	req := httptest.NewRequest(http.MethodPost, "/admin/deductions/personal", strings.NewReader(reqBodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NotEmpty(t, c)

	err := SetPersonalAllowanceAmount(c)

	require.Error(t, err)
}

func TestSetPersonalAllowanceAmountWithInvalidAmount(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	e := echo.New()
	requestBody := Allowances{
		Amount: 100001,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE allowances SET personal = $1 WHERE id = $2`)).
		WithArgs(requestBody.Amount, 1).WillReturnResult(sqlmock.NewResult(1, 1))

	rec, c := mockNewRequestAdmin(requestBody, t, e, "/admin/deductions/personal")

	_ = SetPersonalAllowanceAmount(c)

	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetPersonalAllowanceAmountButQueryError(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	e := echo.New()
	requestBody := Allowances{
		Amount: 100000,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE allowances SET personal = $1 WHERE id = $2`)).
		WithoutArgs().WillReturnError(sql.ErrNoRows)

	_, c := mockNewRequestAdmin(requestBody, t, e, "/admin/deductions/personal")

	err := SetPersonalAllowanceAmount(c)

	require.Error(t, err)
}

func TestSetKReceiptAllowanceAmount(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	e := echo.New()
	requestBody := Allowances{
		Amount: 70000,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE allowances SET "k-receipt" = $1 WHERE id = $2`)).
		WithArgs(requestBody.Amount, 1).WillReturnResult(sqlmock.NewResult(1, 1))

	rec, c := mockNewRequestAdmin(requestBody, t, e, "/admin/deductions/k-receipt")

	err := SetKReceiptAllowanceAmount(c)

	require.NoError(t, err)
	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSetKReceiptAllowanceAmounttWithInvalidRequest(t *testing.T) {
	db, _ := setupMockDB()
	conn = db

	e := echo.New()
	reqBodyJSON := `{"amount": "not a number"}`

	req := httptest.NewRequest(http.MethodPost, "/admin/deductions/k-receipt", strings.NewReader(reqBodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NotEmpty(t, c)

	err := SetKReceiptAllowanceAmount(c)

	require.Error(t, err)
}

func TestSetKReceiptAmountWithInvalidAmount(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	e := echo.New()
	requestBody := Allowances{
		Amount: 100001,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE allowances SET "k-receipt" = $1 WHERE id = $2`)).
		WithArgs(requestBody.Amount, 1).WillReturnResult(sqlmock.NewResult(1, 1))

	rec, c := mockNewRequestAdmin(requestBody, t, e, "/admin/deductions/personal")

	_ = SetKReceiptAllowanceAmount(c)

	require.NotEmpty(t, rec.Body)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetKReceiptAllowanceAmountButQueryError(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	e := echo.New()
	requestBody := Allowances{
		Amount: 100000,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE allowances SET "k-receipt" = $1 WHERE id = $2`)).
		WithoutArgs().WillReturnError(sql.ErrNoRows)

	_, c := mockNewRequestAdmin(requestBody, t, e, "/admin/deductions/personal")

	err := SetKReceiptAllowanceAmount(c)

	require.Error(t, err)
}

func mockNewRequestAdmin(requestBody Allowances, t *testing.T, e *echo.Echo, url string) (*httptest.ResponseRecorder, echo.Context) {
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
