package database

import (
	"database/sql"
	"os"
)

var database *sql.DB

// InitializeDatabase ...
// Initializes database connection
// for API
func InitializeDatabase() {
	connectionString := os.Getenv("DATABASE_URL")

	if connectionString == "" {
		connectionString = "postgres://postgres:root@localhost:5432/finlit?sslmode=disable"
	}

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		panic(err)
	}

	database = db
}

// GetConnection ...
// Returns a transaction valid for
// 1 db connection.
// Once db interaction is complete
// call tx.Commit() to return the
// connection to the pool
func GetConnection() *sql.Tx {
	tx, err := database.Begin()

	if err != nil {
		panic(err)
	}

	return tx
}

// PrepareStatement ...
// Prepares statement and
// returns executable stmt object
func PrepareStatement(tx *sql.Tx, query string) *sql.Stmt {
	stmt, err := tx.Prepare(query)
	if err != nil {
		panic(err)
	}

	return stmt
}
