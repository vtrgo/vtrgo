package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Recipient represents an email user
type User struct {
	Name   string
	Email  string
	Active bool
}

func InitUserDB(db *sql.DB) error {
	recipientQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		active BOOLEAN NOT NULL
	);
	`

	_, err := db.Exec(recipientQuery)
	return err
}

func InsertUser(db *sql.DB, name, email string, active bool) error {
	query := `INSERT INTO users (name, email, active) VALUES (?, ?, ?)`
	_, err := db.Exec(query, name, email, active)
	return err
}

func RemoveUser(db *sql.DB, email string) error {
	query := `DELETE FROM users WHERE email = ?`
	_, err := db.Exec(query, email)
	return err
}

func FetchUsers(db *sql.DB, activeOnly bool) ([]User, error) {
	var query string
	if activeOnly {
		query = `SELECT name, email, active FROM users WHERE active = 1`
	} else {
		query = `SELECT name, email, active FROM users`
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Name, &user.Email, &user.Active); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
