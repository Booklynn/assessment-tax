package tax

import (
	"database/sql"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetPersonalAllowanceValid(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(60000)
	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	got, err := getPersonalAllowance()

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, 60000.0, got)
}

func TestGetPersonalAllowanceReturnError(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	mock.ExpectQuery("SELECT personal FROM allowances WHERE id = ?").WithArgs(1).WillReturnError(sql.ErrNoRows)

	got, err := getPersonalAllowance()

	require.Empty(t, got)
	require.EqualError(t, err, "no record found with the specified id")
}

func TestGetDonationAllowanceValid(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(100000)
	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnRows(rows)

	got, err := getDonationAllowance()

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, 100000.0, got)
}

func TestGetDonationAllowanceReturnError(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	mock.ExpectQuery("SELECT donation FROM allowances WHERE id = ?").WithArgs(1).WillReturnError(sql.ErrNoRows)

	got, err := getDonationAllowance()

	require.Empty(t, got)
	require.EqualError(t, err, "no record found with the specified id")
}

func TestGetKReceiptAllowanceValid(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	rows := mock.NewRows([]string{"personal"}).AddRow(50000)
	mock.ExpectQuery(`SELECT "k-receipt" FROM allowances WHERE id = ?`).WithArgs(1).WillReturnRows(rows)

	got, err := getKReceiptAllowance()

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, 50000.0, got)
}

func TestGetKReceiptAllowanceReturnError(t *testing.T) {
	db, mock := setupMockDB()
	conn = db

	mock.ExpectQuery("SELECT k-receipt FROM allowances WHERE id = ?").WithArgs(1).WillReturnError(sql.ErrNoRows)

	got, err := getKReceiptAllowance()

	require.Empty(t, got)
	require.EqualError(t, err, "no record found with the specified id")
}

func setupMockDB() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal("Cannot connect to database.", err)
	}
	return db, mock
}
