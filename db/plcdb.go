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

func InitDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		length INTEGER
	);
	`
	_, err := db.Exec(query)
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
