package db

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Alarm represents the structure of the XML file
type Alarm struct {
	Triggers []Trigger `xml:"triggers>trigger"`
	Messages []Message `xml:"messages>message"`
}

// Trigger represents a trigger in the XML file
type Trigger struct {
	ID    string `xml:"id,attr"`
	Label string `xml:"label,attr"`
}

// Message represents a message in the XML file
type Message struct {
	ID      string `xml:"id,attr"`
	Trigger string `xml:"trigger,attr"`
	Text    string `xml:"text,attr"`
}

func InitAlarmDB(db *sql.DB) error {
	alarmQuery := `
	CREATE TABLE IF NOT EXISTS alarms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trigger TEXT NOT NULL UNIQUE,
		message TEXT NOT NULL,
		tag TEXT NOT NULL
	);
	`
	_, err := db.Exec(alarmQuery)
	if err != nil {
		return err
	}
	return err
}

func InsertAlarms(db *sql.DB, alarms Alarm) error {
	for _, trigger := range alarms.Triggers {
		var messageText string
		for _, message := range alarms.Messages {
			if message.Trigger == "#"+trigger.ID {
				messageText = message.Text
				break
			}
		}

		if messageText != "" {
			_, err := db.Exec("INSERT INTO alarms (trigger, message, tag) VALUES (?, ?, ?)", trigger.ID, messageText, trigger.Label)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkAlarms(db *sql.DB) error {
	rows, err := db.Query("SELECT id, trigger, message, tag FROM alarms")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var trigger, message, tag string
		err = rows.Scan(&id, &trigger, &message, &tag)
		if err != nil {
			return err
		}
		log.Printf("ID: %d, Trigger: %s, Message: %s, Tag: %s\n", id, trigger, message, tag)
		fmt.Printf("ID: %d, Trigger: %s, Message: %s, Tag: %s\n", id, trigger, message, tag)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func Work() {

	xmlFile, err := os.Open("resources/22-045 Halkey FTAlarms for VSCode.xml")
	if err != nil {
		fmt.Println("Error opening XML file:", err)
		return
	}
	defer xmlFile.Close()

	byteValue, _ := io.ReadAll(xmlFile)

	var alarms Alarm
	alarm_list := xml.Unmarshal(byteValue, &alarms)
	log.Println(alarm_list)

	db, err := sql.Open("sqlite3", "./alarmsdb.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	err = InitAlarmDB(db)
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	// err = InsertAlarms(db, alarms)
	// if err != nil {
	// 	fmt.Println("Error inserting alarms into database:", err)
	// 	return
	// }

	// fmt.Println("Alarms inserted successfully")

	err = checkAlarms(db)
	if err != nil {
		fmt.Println("Error checking alarms in database:", err)
		return
	}

	fmt.Println("Alarms listed successfully")

}
