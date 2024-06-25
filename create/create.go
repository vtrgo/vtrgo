package create

import (
	"os"
	"text/template"
	"time"
)

type PLCTag struct {
	Name        string
	TagType     string
	DataType    string
	Description string
	L5KData     string
	Value       string
}

type L5XData struct {
	TargetName           string
	ExportDate           string
	ControllerName       string
	ProcessorType        string
	MajorRev             string
	MinorRev             string
	TimeSlice            string
	ShareUnusedTimeSlice string
	ProjectCreationDate  string
	LastModifiedDate     string
	ProjectSN            string
	CatalogNumber        string
	Tags                 []PLCTag
}

func L5XCreate() {
	tmpl, err := template.ParseFiles("create/template.l5x")
	if err != nil {
		panic(err)
	}

	data := L5XData{
		TargetName:           "ImportedFromL5X",
		ExportDate:           time.Now().Format(time.RFC1123),
		ControllerName:       "ImportedFromL5X",
		ProcessorType:        "1769-L33ER",
		MajorRev:             "35",
		MinorRev:             "11",
		TimeSlice:            "20",
		ShareUnusedTimeSlice: "1",
		ProjectCreationDate:  time.Now().Add(-2 * time.Minute).Format(time.RFC1123),
		LastModifiedDate:     time.Now().Format(time.RFC1123),
		ProjectSN:            "16#0000_0000",
		CatalogNumber:        "1769-L33ER",
		Tags: []PLCTag{
			{Name: "MotorOn", TagType: "Base", DataType: "BOOL", Description: "Motor Is Running Bit", L5KData: "0", Value: "0"},
			{Name: "MotorSpeed", TagType: "Base", DataType: "DINT", Description: "Motor Speed", L5KData: "0", Value: "0"},
		},
	}

	file, err := os.Create("output.L5X")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		panic(err)
	}
}
