package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// ConfigVariable represents a configuration variable stored in the database
type ConfigVariable struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// InitConfigDB initializes the configuration database
func InitConfigDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// InsertConfigVariable inserts a new configuration variable into the database
func InsertConfigVariable(db *sql.DB, name, value string) error {
	query := `INSERT INTO config (name, value) VALUES (?, ?)`
	_, err := db.Exec(query, name, value)
	return err
}

// RemoveConfigVariable removes a configuration variable from the database by name
func RemoveConfigVariable(db *sql.DB, name string) error {
	query := `DELETE FROM config WHERE name = ?`
	_, err := db.Exec(query, name)
	return err
}

// FetchConfigVariables retrieves all configuration variables from the database
func FetchConfigVariables(db *sql.DB) ([]ConfigVariable, error) {
	rows, err := db.Query(`SELECT id, name, value FROM config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variables []ConfigVariable
	for rows.Next() {
		var variable ConfigVariable
		if err := rows.Scan(&variable.ID, &variable.Name, &variable.Value); err != nil {
			return nil, err
		}
		variables = append(variables, variable)
	}
	return variables, nil
}
