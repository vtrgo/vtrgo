package tagdb

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// PlcTag represents a PLC tag that is to be stored in the database
type PlcTag struct {
	Name   string
	Type   string
	Value  interface{}
	Length int
}

// Initialize the database and create the table
func InitTagDB(db *sql.DB, table string) error {
	tagTableQuery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		length INTEGER
	);`, table)

	_, err := db.Exec(tagTableQuery)
	if err != nil {
		return err
	}
	return err
}

func InsertTag(db *sql.DB, table string, name, tagType string, length int) error {
	query := fmt.Sprintf(`INSERT INTO %s (name, type, length) VALUES (?, ?, ?)`, table)
	_, err := db.Exec(query, name, tagType, length)
	return err
}

func RemoveTag(db *sql.DB, table string, name string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE name = ?`, table)
	_, err := db.Exec(query, name)
	return err
}

func FetchTags(db *sql.DB, table string) ([]PlcTag, error) {
	query := fmt.Sprintf(`SELECT name, type, length FROM %s`, table)
	rows, err := db.Query(query)
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
