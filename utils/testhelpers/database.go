package testhelpers

import (
	"database/sql"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
)

// GetTestDbTransaction returns a mock db transaction instance for unit testing
func GetTestDbTransaction() (*sql.Tx, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()

	if err != nil {
		fmt.Println("Failed to initialize mock db for unit test")
		panic(err)
	}

	mock.ExpectBegin()

	tx, err := db.Begin()

	if err != nil {
		fmt.Println("Error occurred when beginning transaction in TestDbCommit unit test")
		panic(err)
	}

	return tx, mock
}
