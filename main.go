package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	plcdb "vtrgo/db"
	"vtrgo/excel"
	"vtrgo/plc"

	_ "github.com/mattn/go-sqlite3"
)

var tagdb *sql.DB

func main() {

	// Test a command line introduction to the user
	welcome("Justin")

	// Declare a PLC tag as a test integer variable
	myTag := plcdb.PlcTag{
		Name:  "Program:HMI_Executive_Control.TestDint",
		Type:  "int32",
		Value: 0}

	// Create a new PLC identity using the IP address of the Logix controller
	plc := plc.NewPLC("10.103.115.10")

	// Make a connection to the PLC
	err := plc.Connect()
	if err != nil {
		log.Printf("Error connecting to PLC: %v", err)
		return
	}
	defer plc.Disconnect()

	// Test read an integer value from the PLC
	myTag.Value, err = plc.ReadTagInt32(myTag.Name)
	if err != nil {
		log.Printf("Error reading tag: %v", err)
		return
	}
	fmt.Printf("Tag was value: %d\n", myTag.Value)

	// Write an integer test value to the PLC
	myTag.Value = myTag.Value.(int32) + 1
	err = plc.WriteTagInt32(myTag.Name, myTag.Value.(int32))
	if err != nil {
		log.Printf("Error writing to tag: %v", err)
		return
	}
	fmt.Println("Tag value changed to:", myTag.Value)

	// Create a connection to the plc_tags database
	tagdb, err = sql.Open("sqlite3", "./plc_tags.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer tagdb.Close()

	// Initialize and create the database if it does not already exist
	err = plcdb.InitDB(tagdb)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Creates the /metrics endpoint to display and update tag values in the browser
	http.HandleFunc("/metrics", func(write http.ResponseWriter, read *http.Request) {
		myTag.Value, err = plc.ReadTagInt32(myTag.Name)
		htmlResponse := fmt.Sprintf("<p>%s: <span id=\"tagValue\">%d</span></p>", myTag.Name, myTag.Value)
		write.Header().Set("Content-Type", "text/html")
		write.Write([]byte(htmlResponse))
	})

	// Creates the /update endpoint to increment the test integer value and store the new value in excel
	http.HandleFunc("/update", func(write http.ResponseWriter, read *http.Request) {
		myTag.Value = myTag.Value.(int32) + 1
		plc.WriteTagInt32(myTag.Name, myTag.Value.(int32))
		log.Printf("Tag value updated: %v", myTag.Value)

		htmlResponse := fmt.Sprintf("<p>%s: <span id=\"tagValue\">%d</span></p>", myTag.Name, myTag.Value)
		write.Header().Set("Content-Type", "text/html")
		write.Write([]byte(htmlResponse))

		plcData := excel.PlcTags{
			Tags: []excel.Tag{
				{Name: myTag.Name, Value: myTag.Value.(int32)},
			},
		}

		err = excel.WriteDataToExcel(plcData, "plc_data.xlsx")
		if err != nil {
			log.Printf("Failed to write to Excel: %v", err)
			http.Error(write, "Failed to write to Excel", http.StatusInternalServerError)
		}
	})

	triggerTag := "Program:HMI_Executive_Control.DataTrigger"
	responseTag := "Program:HMI_Executive_Control.TriggerResponse"

	interval := 3 * time.Millisecond

	date := time.Now().Format("2006-01-02")
	customerName := "Halkey"
	recipeTag := "HMI_Recipe[0].RecipeName"
	recipeName, err := plc.ReadTagString(recipeTag)
	if err != nil {
		log.Printf("Error reading tag: %v", err)
		return
	}
	filePath := fmt.Sprintf("%s_%s-Data_%s.xlsx", customerName, recipeName, date)

	startTriggerChecker(tagdb, plc, triggerTag, responseTag, filePath, interval)

	// Sets up endpoint handlers for each function call
	http.HandleFunc("/add-tag", addTagHandler)
	http.HandleFunc("/remove-tag", removeTagHandler)
	http.HandleFunc("/list-tags", listTagsHandler)
	http.HandleFunc("/load-list", loadListHandler)
	http.HandleFunc("/load-add", loadAddHandler)
	http.HandleFunc("/load-remove", loadRemoveHandler)

	fs := http.FileServer(http.Dir("."))
	http.Handle("resource/styles.css", fs)
	// http.Handle("/list-tags.html", fs)
	// http.Handle("/add-tags.html", fs)
	// http.Handle("/remove-tags.html", fs)

	http.Handle("/", fs)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	// create.L5XCreate()
}

func loadListHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "list-tags.html")
}

func loadAddHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "add-tags.html")
}

func loadRemoveHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "remove-tags.html")
}

// Handles the /add-tag endpoint for adding new tags to the plc_tags database
func addTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("tag")
		tagType := r.FormValue("type")
		if tag == "" || tagType == "" {
			http.Error(w, "Tag name and type are required", http.StatusBadRequest)
			return
		}

		err := plcdb.InsertTag(tagdb, tag, tagType)
		if err != nil {
			log.Printf("Failed to insert tag: %v", err)
			http.Error(w, "Failed to insert tag", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Tag '%s' with type '%s' added successfully!", tag, tagType)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Handles the /remove-tag endpoint for deleting tags from the plc_tags database
func removeTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("removed-tag")
		if tag == "" {
			http.Error(w, "Tag name required", http.StatusBadRequest)
			return
		}

		err := plcdb.RemoveTag(tagdb, tag)
		if err != nil {
			log.Printf("Failed to remove tag: %v", err)
			http.Error(w, "Failed to delete tag", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Tag '%s' removed successfully!", tag)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Handles the /list-tags endpoint for displaying all the tags stored in the plc_tags database
func listTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := plcdb.FetchTags(tagdb)
	if err != nil {
		log.Printf("Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<ul>")
	for _, tag := range tags {
		fmt.Fprintf(w, "<li>%s (%s)</li>", tag.Name, tag.Type)
	}
	fmt.Fprintf(w, "</ul>")
}

// Goroutine to monitor a boolean trigger tag in the PLC.
// When triggerTag is activated (True state), reads all tag values in the plc_tags database and stores them in excel
func startTriggerChecker(tagdb *sql.DB, plc *plc.PLC, triggerTag string, responseTag string, filePath string, interval time.Duration) {

	go func() {

		for {
			trigger, err := plc.ReadTagBool(triggerTag)
			if err != nil {
				log.Printf("Error checking trigger: %v", err)
				time.Sleep(interval)
				continue
			}

			if trigger {
				log.Println("Trigger activated, writing data to Excel")

				tags, err := plcdb.FetchTags(tagdb)
				if err != nil {
					log.Printf("Failed to fetch tags: %v", err)
					time.Sleep(interval)
					continue
				}

				var plcTags []excel.Tag
				for _, tag := range tags {
					var tagValue interface{}
					var err error

					switch tag.Type {
					case "int32":
						tagValue, err = plc.ReadTagInt32(tag.Name)
					case "real":
						tagValue, err = plc.ReadTagFloat32(tag.Name)
					case "bool":
						tagValue, err = plc.ReadTagBool(tag.Name)
					case "string":
						tagValue, err = plc.ReadTagString(tag.Name)

					default:
						log.Printf("Unknown tag type %s for tag %s", tag.Type, tag.Name)
						continue
					}

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
				err = plc.WriteTagInt32(responseTag, 1)
				if err != nil {
					log.Printf("Failed to write response tag: %v", err)
				}
			}

			time.Sleep(interval)
		}
	}()
}
