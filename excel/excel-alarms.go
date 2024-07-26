package excel

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

// Tag represents a PLC tag with a name and value
type AlarmTag struct {
	Name    string
	Message string
	Value   interface{}
}

// PlcTags represents a collection of PLC tags
type AlarmTags struct {
	AlarmTags []AlarmTag
}

func WriteAlarmsToExcel(data AlarmTags, filePath string) error {
	var file *excelize.File
	var err error

	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		// Create a new Excel file if it doesn't exist
		file = excelize.NewFile()
		// Create a new sheet
		index, _ := file.NewSheet("Sheet1")
		// Set the value of headers
		file.SetCellValue("Sheet1", "A1", "Timestamp")
		file.SetCellValue("Sheet1", "B1", "Alarm Tag")
		file.SetCellValue("Sheet1", "C1", "Alarm Description")
		file.SetCellValue("Sheet1", "D1", "Value")
		file.SetActiveSheet(index)
	} else {
		// Open the existing Excel file
		file, err = excelize.OpenFile(filePath)
		if err != nil {
			return err
		}
	}

	// Find the next empty row
	rows, err := file.GetRows("Sheet1")
	if err != nil {
		return err
	}
	nextRow := len(rows) + 1

	// Get the current timestamp
	timestamp := time.Now().Format(time.RFC3339)

	for _, tag := range data.AlarmTags {
		log.Printf("TAG:%v", tag)
		// Write the metrics data to the next row
		file.SetCellValue("Sheet1", fmt.Sprintf("A%d", nextRow), timestamp)
		file.SetCellValue("Sheet1", fmt.Sprintf("B%d", nextRow), tag.Name)
		file.SetCellValue("Sheet1", fmt.Sprintf("C%d", nextRow), tag.Message)
		file.SetCellValue("Sheet1", fmt.Sprintf("D%d", nextRow), true)
		nextRow++

	}

	// Save the file
	err = file.SaveAs(filePath)
	if err != nil {
		return err
	}

	return nil
}
