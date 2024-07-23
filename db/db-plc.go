package plcdb

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// PlcTag represents a PLC tag that is to be stored in the database
type PlcTag struct {
	Name   string
	Type   string
	Value  interface{}
	Length int
}

// Recipient represents an email user
type User struct {
	Name   string
	Email  string
	Active bool
}

func InitPlcDB(db *sql.DB) error {
	tagQuery := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		length INTEGER
	);
	`
	recipientQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		active BOOLEAN NOT NULL
	);
	`

	_, err := db.Exec(tagQuery)
	if err != nil {
		return err
	}

	_, err = db.Exec(recipientQuery)
	return err
}

func InsertTag(db *sql.DB, name, tagType string, length int) error {
	query := `INSERT INTO tags (name, type, length) VALUES (?, ?, ?)`
	_, err := db.Exec(query, name, tagType, length)
	return err
}

func RemoveTag(db *sql.DB, name string) error {
	query := `DELETE FROM tags WHERE name = ?`
	_, err := db.Exec(query, name)
	return err
}

func FetchTags(db *sql.DB) ([]PlcTag, error) {
	rows, err := db.Query(`SELECT name, type, length FROM tags`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []PlcTag
	for rows.Next() {
		var tag PlcTag
		if err := rows.Scan(&tag.Name, &tag.Type, &tag.Length); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
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
