package main

import (
	"database/sql"
	"log"
	"time"

	plcdb "vtrgo/db"
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

				tags, err := plcdb.FetchTags(tagdb)
				if err != nil {
					log.Printf("Failed to fetch tags: %v", err)
					time.Sleep(interval)
					continue
				}

				var plcTags []excel.Tag
				for _, tag := range tags {
					tagValue, err := plc.ReadTag(tag.Name, tag.Type, tag.Length)

					if err != nil {
						log.Printf("Failed to read Tag %s: %v", tag.Name, err)
						continue
					}

					plcTags = append(plcTags, excel.Tag{Name: tag.Name, Value: tagValue})

				}

				if len(plcTags) > 0 {
					plcData := excel.PlcTags{Tags: plcTags}
					err = excel.WriteDataToExcel(plcData, filePath)
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
