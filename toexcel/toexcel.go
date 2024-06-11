package toexcel

import (
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

// Tag represents a PLC tag with a name and value
type Tag struct {
	Name  string
	Value interface{}
}

// PlcTags represents a collection of PLC tags
type PlcTags struct {
	Tags []Tag
}

// WriteDataToExcel writes the PlcTags to an Excel file
func WriteDataToExcel(data PlcTags, filePath string) error {
	var file *excelize.File
	var err error

	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		// Create a new Excel file if it doesn't exist
		file = excelize.NewFile()
		// Create a new sheet
		index, _ := file.NewSheet("Sheet1")
		// Set the value of headers
		file.SetCellValue("Sheet1", "A1", "Timestamp")
		for i, tag := range data.Tags {
			col := string(rune('B' + i))
			file.SetCellValue("Sheet1", fmt.Sprintf("%s1", col), tag.Name)
		}
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

	// Write the metrics data to the next row
	timestamp := time.Now().Format(time.RFC3339)
	file.SetCellValue("Sheet1", fmt.Sprintf("A%d", nextRow), timestamp)
	for i, tag := range data.Tags {
		col := string(rune('B' + i))
		file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, nextRow), tag.Value)
	}

	// Save the file
	err = file.SaveAs(filePath)
	if err != nil {
		return err
	}

	return nil
}
