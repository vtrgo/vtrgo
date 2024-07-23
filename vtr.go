package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	plcdb "vtrgo/db"
	"vtrgo/tagdb"

	"vtrgo/excel"
	"vtrgo/plc"

	_ "github.com/mattn/go-sqlite3"
)

var dataTagsDb *sql.DB
var alarmTagsDb *sql.DB

type MetricsData struct {
	Value int32 `json:"value"`
}

func main() {

	plcdb.Work()

	// config, err := email.LoadConfig()
	// if err != nil {
	// 	log.Fatal("Error loading config:", err)
	// }
	// recipient := "justin@vtrfeedersolutions.com"
	// subject := "This is an automated message from vtrgo."
	// message := "Please find your data report attached."
	// attachment := "output_files/Halkey_43BK-730-Data_2024-06-10.xlsx"

	// err = email.SendEmail(config, recipient, subject, message, attachment, true)
	// if err != nil {
	// 	log.Println("Error sending email:", err)
	// }

	// log.Println("Email sent successfully!")

	// config, err := email.LoadConfig()
	// if err != nil {
	// 	log.Fatal("Error loading config:", err)
	// }

	// err = email.SendEmail(config, recipient, subject, message, attachment)
	// if err != nil {
	// 	log.Println("Error sending email:", err)
	// } else {
	// 	log.Printf("Data successfully sent to: %v", recipient)
	// }

	// Test a command line introduction to the user
	welcome("Justin")

	// Declare a PLC tag as a test integer variable
	myTag := plcdb.PlcTag{
		Name: "Program:HMI_Executive_Control.TestDint",
		Type: "dint",
	}

	myDintArray := plcdb.PlcTag{
		Name:   "Program:HMI_Executive_Control.RealData",
		Type:   "[]real",
		Length: 10,
	}
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
	myTag.Value, err = plc.ReadTag(myTag.Name, myTag.Type, myTag.Length)
	if err != nil {
		log.Printf("Error reading tag: %v", err)
		return
	}
	fmt.Printf("Tag was value: %d\n", myTag.Value)

	// Write an integer test value to the PLC
	myTag.Value = myTag.Value.(int32) + 1
	newValue, err := plc.WriteTag(myTag.Name, myTag.Type, myTag.Value)
	if err != nil {
		log.Printf("Error writing to tag: %v", err)
		return
	}
	fmt.Println("Tag value changed to:", newValue)

	myDintArray.Value, err = plc.ReadTag(myDintArray.Name, myDintArray.Type, myDintArray.Length)
	if err != nil {
		log.Printf("Error reading tag: %v", err)
		return
	}
	fmt.Printf("Tag: %v has value: %f\n", myDintArray.Name, myDintArray.Value)

	// Type assert the interface{} to []float64
	if values, ok := myDintArray.Value.([]float32); ok {
		for i, v := range values {
			fmt.Printf("Tag: %s[%d] has value: %f\n", myDintArray.Name, i, v)
		}
	} else {
		fmt.Println("Error: Value is not of correct type")
	}

	// Create a connection to the plc_tags database
	dataTagsDb, err = sql.Open("sqlite3", "./plc_tags.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dataTagsDb.Close()

	// Initialize and create the database if it does not already exist
	err = plcdb.InitPlcDB(dataTagsDb)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a connection to the generic tags database, passing the db and table names
	alarmTagsDb, err = sql.Open("sqlite3", "./db_tags.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer alarmTagsDb.Close()

	// Initialize and create the database if it does not already exist
	tableName := "alarmTags"
	err = tagdb.InitTagDB(alarmTagsDb, tableName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	tagdb.FetchTags(alarmTagsDb, tableName)
	// Creates the /metrics endpoint to display and update tag values in the browser
	http.HandleFunc("/metrics", func(write http.ResponseWriter, read *http.Request) {
		myTag.Value, err = plc.ReadTag(myTag.Name, myTag.Type, myTag.Length)
		htmlResponse := fmt.Sprintf("<p>%s: <span id=\"tagValue\">%d</span></p>", myTag.Name, myTag.Value)
		write.Header().Set("Content-Type", "text/html")
		write.Write([]byte(htmlResponse))
	})

	// Creates the /update endpoint to increment the test integer value and store the new value in excel
	http.HandleFunc("/update", func(write http.ResponseWriter, read *http.Request) {
		myTag.Value = myTag.Value.(int32) + 1
		plc.WriteTag(myTag.Name, myTag.Type, myTag.Value)
		log.Printf("Tag value updated: %v", myTag.Value)

		htmlResponse := fmt.Sprintf("<p>%s: <span id=\"tagValue\">%d</span></p>", myTag.Name, myTag.Value)
		write.Header().Set("Content-Type", "text/html")
		write.Write([]byte(htmlResponse))

		plcData := excel.PlcTags{
			Tags: []excel.Tag{
				{Name: myTag.Name, Value: myTag.Value.(int32)},
			},
		}

		err = excel.WriteDataToExcel(plcData, "output_files/plc_data.xlsx")
		if err != nil {
			log.Printf("Failed to write to Excel: %v", err)
			http.Error(write, "Failed to write to Excel", http.StatusInternalServerError)
		}
	})

	// Creates the /update endpoint to increment the test integer value and store the new value in excel
	http.HandleFunc("/metrics/latest", func(write http.ResponseWriter, read *http.Request) {
		// Prepare the response data

		// Convert the response data to JSON
		metricsData := MetricsData{myTag.Value.(int32)}
		jsonResponse, err := json.Marshal(metricsData)
		if err != nil {
			http.Error(write, "Failed to encode JSON response", http.StatusInternalServerError)
			return
		}

		// Set the response headers and write the response
		write.Header().Set("Content-Type", "application/json")
		write.Write(jsonResponse)
	})

	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	tmpl := template.Must(template.ParseFiles("index.html"))
	// 	tmpl.Execute(w, nil)
	// })

	http.HandleFunc("/load-data-tags-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/data-tags-section.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-alarm-tags-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/alarm-tags-section.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-email-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/users-section.html"))
		tmpl.Execute(w, nil)
	})

	triggerTag := "Program:HMI_Executive_Control.DataTrigger"
	responseTag := "Program:HMI_Executive_Control.DataResponse"

	interval := 300 * time.Millisecond

	date := time.Now().Format("2006-01-02")
	customerName := "Halkey"
	recipeTag := "HMI_Recipe[0].RecipeName"
	recipeName, err := plc.ReadTagString(recipeTag)
	if err != nil {
		log.Printf("Error reading tag: %v", err)
		return
	}
	filePath := fmt.Sprintf("output_files/%s_%s-Data_%s.xlsx", customerName, recipeName, date)

	startTriggerChecker(dataTagsDb, plc, triggerTag, responseTag, filePath, interval)

	// Sets up endpoint handlers for each function call
	http.HandleFunc("/add-tag", addTagHandler)
	http.HandleFunc("/remove-tag", removeTagHandler)
	http.HandleFunc("/list-tags", listTagsHandler)
	http.HandleFunc("/add-alarm-tag", addAlarmTagHandler)
	http.HandleFunc("/remove-alarm-tag", removeAlarmTagHandler)
	http.HandleFunc("/list-alarm-tags", listAlarmTagsHandler)
	http.HandleFunc("/list-remove-tags", listRemoveTagsHandler)
	http.HandleFunc("/load-list-tags", loadListTagsHandler)
	http.HandleFunc("/load-add-tags", loadAddTagsHandler)
	http.HandleFunc("/load-remove-tags", loadRemoveTagsHandler)
	http.HandleFunc("/load-list-alarm-tags", loadListAlarmTagsHandler)
	http.HandleFunc("/load-add-alarm-tags", loadAddAlarmTagsHandler)
	http.HandleFunc("/load-remove-alarm-tags", loadRemoveAlarmTagsHandler)
	http.HandleFunc("/add-user", addUserHandler)
	http.HandleFunc("/remove-user", removeUserHandler)
	http.HandleFunc("/list-users", listUsersHandler)
	http.HandleFunc("/load-list-users", loadListUsersHandler)
	http.HandleFunc("/load-add-users", loadAddUsersHandler)
	http.HandleFunc("/load-remove-users", loadRemoveUsersHandler)
	http.HandleFunc("/load-list-remove-tags", loadListRemoveTagsHandler)
	http.HandleFunc("js/metricsChart.js", metricsChartHandler)

	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	// create.L5XCreate()

}

func metricsChartHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "resources/metricsChart.js")
}

func loadListTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/list-tags.html")
}

func loadAddTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/add-tags.html")
}

func loadRemoveTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/remove-tags.html")
}

func loadListAlarmTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/list-alarm-tags.html")
}

func loadAddAlarmTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/add-alarm-tags.html")
}

func loadRemoveAlarmTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/remove-alarm-tags.html")
}

func loadListUsersHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/list-users.html")
}

func loadAddUsersHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/add-users.html")
}

func loadRemoveUsersHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/remove-users.html")
}

func loadListRemoveTagsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/list-remove-tags.html")
}

// Handles the /list-tags endpoint for displaying all the tags stored in the plc_tags database
func listTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := plcdb.FetchTags(dataTagsDb)
	if err != nil {
		log.Printf("Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<ul>")
	for _, tag := range tags {
		fmt.Fprintf(w, "<li>%s (%s)[%d]</li>", tag.Name, tag.Type, tag.Length)
	}
	fmt.Fprintf(w, "</ul>")
}

// Handles the /add-tag endpoint for adding new tags to the plc_tags database
func addTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("tag")
		tagType := r.FormValue("type")
		lengthText := r.FormValue("length")
		length, err := strconv.Atoi(lengthText)
		if tag == "" || tagType == "" {
			http.Error(w, "Tag name and type are required", http.StatusBadRequest)
			return
		}

		plcdb.InsertTag(dataTagsDb, tag, tagType, length)
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

		err := plcdb.RemoveTag(dataTagsDb, tag)
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
func listAlarmTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := tagdb.FetchTags(alarmTagsDb, "alarmTags")
	if err != nil {
		log.Printf("Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<ul>")
	for _, tag := range tags {
		fmt.Fprintf(w, "<li>%s (%s)[%d]</li>", tag.Name, tag.Type, tag.Length)
	}
	fmt.Fprintf(w, "</ul>")
}

// Handles the /add-tag endpoint for adding new tags to the plc_tags database
func addAlarmTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("tag")
		tagType := r.FormValue("type")
		lengthText := r.FormValue("length")
		length, err := strconv.Atoi(lengthText)
		if tag == "" || tagType == "" {
			http.Error(w, "Tag name and type are required", http.StatusBadRequest)
			return
		}

		tagdb.InsertTag(alarmTagsDb, "alarmTags", tag, tagType, length)
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
func removeAlarmTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("removed-tag")
		if tag == "" {
			http.Error(w, "Tag name required", http.StatusBadRequest)
			return
		}

		err := tagdb.RemoveTag(alarmTagsDb, "alarmTags", tag)
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
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := plcdb.FetchUsers(dataTagsDb, false)
	if err != nil {
		log.Printf("Failed to fetch users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<ul>")
	for _, user := range users {
		fmt.Fprintf(w, "<li>%s (%s)[%v]</li>", user.Name, user.Email, user.Active)
	}
	fmt.Fprintf(w, "</ul>")
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := r.FormValue("name")
		email := r.FormValue("email")
		active := r.FormValue("active") == "on"

		if user == "" || email == "" {
			http.Error(w, "User name and email are required", http.StatusBadRequest)
			return
		}

		err := plcdb.InsertUser(dataTagsDb, user, email, active)
		if err != nil {
			log.Printf("Failed to insert user: %v", err)
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Tag '%s' with type '%s' added successfully!", user, email)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Handles the /remove-tag endpoint for deleting tags from the plc_tags database
func removeUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := r.FormValue("removed-user")
		if user == "" {
			http.Error(w, "User name required", http.StatusBadRequest)
			return
		}

		err := plcdb.RemoveUser(dataTagsDb, user)
		if err != nil {
			log.Printf("Failed to remove user: %v", err)
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "User '%s' removed successfully!", user)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Handles the /list-remove-tags endpoint for displaying all the tags in the database with a delete buttin next to each tag
func listRemoveTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := plcdb.FetchTags(dataTagsDb)
	if err != nil {
		log.Printf("Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<ul>")
	for _, tag := range tags {
		fmt.Fprintf(w, `
		<li>
		%s (%s)
		<button hx-delete="/remove-tag?name=%s" hx-swap="innerHTML">Delete</button>
		</li>
		`, tag.Name, tag.Type, tag.Name)
		plcdb.RemoveTag(dataTagsDb, tag.Name)
	}
	fmt.Fprintf(w, "</ul>")
}

// Goroutine to monitor a boolean trigger tag in the PLC.
// When triggerTag is activated (True state), reads all tag values in the plc_tags database and stores them in excel
func startTriggerChecker(dataTagsDb *sql.DB, plc *plc.PLC, triggerTag string, responseTag string, filePath string, interval time.Duration) {

	go func() {

		for {
			trigger, err := plc.ReadTrigger(triggerTag)
			if err != nil {
				log.Printf("Error checking trigger: %v", err)
				time.Sleep(interval)
				continue
			}

			if trigger {
				log.Println("Trigger activated, writing data to Excel")

				tags, err := plcdb.FetchTags(dataTagsDb)
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
