package excel

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
		colIndex := 1
		for _, tag := range data.Tags {
			switch v := tag.Value.(type) {
			case []int32:
				for i := range v {
					col := string(rune('A' + colIndex))
					file.SetCellValue("Sheet1", fmt.Sprintf("%s1", col), fmt.Sprintf("%s[%d]", tag.Name, i))
					colIndex++
				}
			case []float32:
				for i := range v {
					col := string(rune('A' + colIndex))
					file.SetCellValue("Sheet1", fmt.Sprintf("%s1", col), fmt.Sprintf("%s[%d]", tag.Name, i))
					colIndex++
				}

			default:
				col := string(rune('A' + colIndex))
				file.SetCellValue("Sheet1", fmt.Sprintf("%s1", col), tag.Name)
				colIndex++
			}
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
	colIndex := 1
	for _, tag := range data.Tags {
		switch v := tag.Value.(type) {
		case []int:
			for _, val := range v {
				col := string(rune('A' + colIndex))
				file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, nextRow), val)
				colIndex++
			}
		case []int32:
			for _, val := range v {
				col := string(rune('A' + colIndex))
				file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, nextRow), val)
				colIndex++
			}
		case []float32:
			for _, val := range v {
				col := string(rune('A' + colIndex))
				file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, nextRow), val)
				colIndex++
			}
		case []float64:
			for _, val := range v {
				col := string(rune('A' + colIndex))
				file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, nextRow), val)
				colIndex++
			}
		default:
			col := string(rune('A' + colIndex))
			file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, nextRow), tag.Value)
			colIndex++
		}
	}

	// Save the file
	err = file.SaveAs(filePath)
	if err != nil {
		return err
	}

	return nil
}
