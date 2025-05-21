package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"vtrgo/db"
	"vtrgo/excel"
	"vtrgo/plc"

	_ "github.com/mattn/go-sqlite3"
)

// Goroutine to monitor a boolean trigger tag in the PLC.
// When triggerTag is activated (True state), reads all tag values in the plc_tags database and stores them in Excel or InfluxDB
func dataTriggerChecker(
	dataTagsDb *sql.DB,
	plc *plc.PLC,
	triggerTag string,
	responseTag string,
	filePath string,
	interval time.Duration,
	influx *db.InfluxDBClient, // Pass nil if not using InfluxDB
	storeToInflux bool, // true = store to InfluxDB, false = store to Excel
) {

	go func() {

		for {
			trigger, err := plc.ReadTrigger(triggerTag)
			if err != nil {
				log.Printf("Error checking trigger: %v", err)
				time.Sleep(interval)
				continue
			}

			if trigger {

				log.Println("Data trigger activated, capturing data")

				tags, err := db.FetchTags(dataTagsDb, "dataTags")
				if err != nil {
					log.Printf("Failed to fetch tags: %v", err)
					time.Sleep(interval)
					continue
				}

				if storeToInflux && influx != nil {
					for _, tag := range tags {
						tagValue, err := plc.ReadTag(tag.Name, tag.Type, tag.Length)
						if err != nil {
							log.Printf("Failed to read Tag %s: %v", tag.Name, err)
							continue
						}
						fields := map[string]interface{}{"value": tagValue}
						tagsMap := map[string]string{"tag": tag.Name, "type": tag.Type}
						err = influx.WritePLCData(context.Background(), "measurement1", tagsMap, fields, time.Now())
						if err != nil {
							log.Printf("Failed to write tag %s to InfluxDB: %v", tag.Name, err)
						}
					}
				} else {
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
