package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"vtrgo/db"
	"vtrgo/excel"
	"vtrgo/plc"

	_ "github.com/mattn/go-sqlite3"
)

// Goroutine to monitor a boolean trigger tag in the PLC.
// When triggerTag is activated (True state), reads all tag values in the plc_tags database and stores them in excel
func alarmTriggerRoutine(tagdb *sql.DB, plc *plc.PLC, triggerTag string, responseTag string, filePath string, interval time.Duration) {

	go func() {

		for {
			trigger, err := plc.ReadTrigger(triggerTag)
			if err != nil {
				log.Printf("Error checking trigger: %v", err)
				time.Sleep(interval)
				continue
			}

			if trigger {
				log.Println("Alarm trigger activated, loading alarm tags")
				alarmsDb, err := sql.Open("sqlite3", "./alarmsdb.db")
				if err != nil {
					fmt.Println("Error opening database:", err)
					return
				}
				defer alarmsDb.Close()

				err = db.InitAlarmDB(alarmsDb)
				if err != nil {
					fmt.Println("Error initializing database:", err)
					return
				}

				tags, err := db.FetchTags(tagdb, "alarmTags")
				if err != nil {
					log.Printf("Failed to fetch tags: %v", err)
					time.Sleep(interval)
					continue
				}

				var plcTags []excel.AlarmTag
				for _, tag := range tags {
					tagValue, err := plc.ReadTag(tag.Name, tag.Type, tag.Length)

					if err != nil {
						log.Printf("Failed to read Tag %s: %v", tag.Name, err)
						continue
					}

					// Type assertion for []int32
					intTagValues, ok := tagValue.([]int32)
					if !ok {
						log.Printf("Tag %s has an unexpected type: %T", tag.Name, tagValue)
						continue
					}

					// Iterate over each int32 element in the array
					for index, intTagValue := range intTagValues {
						// Check which bits are true and log them
						for i := 0; i < 32; i++ { // int32 has 32 bits
							if (intTagValue & (1 << i)) != 0 {
								tagName := fmt.Sprintf("%s[%d].%d", tag.Name, index, i)
								// Fetch description from the database
								tagMessage, err := db.FetchDescription(alarmsDb, tagName)
								if err != nil {
									log.Printf("Failed to fetch description for %s: %v", tagName, err)
									tagMessage = "Description not found"
								}
								plcTags = append(plcTags, excel.AlarmTag{
									Message: tagMessage,
									Name:    tagName,
									Value:   true,
								})
							}
						}
					}
				}

				if len(plcTags) > 0 {
					// Pass the tags to be written to Excel
					log.Printf("plcTags: %v", plcTags)
					err = excel.WriteAlarmsToExcel(excel.AlarmTags{AlarmTags: plcTags}, filePath)
					if err != nil {
						log.Printf("Failed to write to Excel: %v", err)
					}
				}
				err = plc.WriteResponse(responseTag, true)
				if err != nil {
					log.Printf("Failed to write response tag: %v", err)
				}
			}

			time.Sleep(interval)
		}
	}()

}
