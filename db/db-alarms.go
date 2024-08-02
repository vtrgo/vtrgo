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

type AlarmMessage struct {
	ID      int
	Trigger string
	Message string
	Tag     string
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

func CheckAlarms(db *sql.DB) ([]AlarmMessage, error) {
	rows, err := db.Query("SELECT id, trigger, message, tag FROM alarms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alarms []AlarmMessage

	for rows.Next() {
		var alarm AlarmMessage
		err = rows.Scan(&alarm.ID, &alarm.Trigger, &alarm.Message, &alarm.Tag)
		if err != nil {
			return nil, err
		}
		// log.Printf("ID: %d, Trigger: %s, Message: %s, Tag: %s\n", alarm.ID, alarm.Trigger, alarm.Message, alarm.Tag)
		// fmt.Printf("ID: %d, Trigger: %s, Message: %s, Tag: %s\n", alarm.ID, alarm.Trigger, alarm.Message, alarm.Tag)
		alarms = append(alarms, alarm)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return alarms, nil
}

func listAlarms(db *sql.DB) error {
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

func FetchDescription(db *sql.DB, alarmTag string) (string, error) {
	var description string

	// Execute the query to fetch the description
	err := db.QueryRow("SELECT message FROM alarms WHERE tag = ?", alarmTag).Scan(&description)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows found
			return "", nil
		}
		return "", err
	}

	log.Printf("Message: %s given from Tag: %s", description, alarmTag)
	return description, nil
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

	err = listAlarms(db)
	if err != nil {
		fmt.Println("Error checking alarms in database:", err)
		return
	}

	fmt.Println("Alarms listed successfully")

}
