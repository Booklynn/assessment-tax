package tax

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var conn *sql.DB

func ConnectDb() {
	var err error

	conn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal("Cannot connect to database.", err)
	}
}

func getPersonalAllowance() (float64, error) {
	var personalAllowance float64
	err := conn.QueryRow("SELECT personal FROM allowances WHERE id = $1", 1).Scan(&personalAllowance)
	if err != nil {
		return 0, errors.New("no record found with the specified id")
	}
	return personalAllowance, nil
}
