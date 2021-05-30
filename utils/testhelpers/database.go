package testhelpers

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

// GetTestDbTransaction returns a mock db transaction instance for unit testing
func GetTestDbTransaction() (*sql.Tx, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()

	if err != nil {
		panic(err)
	}

	mock.ExpectBegin()

	tx, err := db.Begin()

	if err != nil {
		panic(err)
	}

	return tx, mock
}
