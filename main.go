// VTRGo
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"vtrgo/db"
	"vtrgo/plc"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var dataTagsDb *sql.DB
var alarmTagsDb *sql.DB
var usersDb *sql.DB

type MetricsData struct {
	Value int32 `json:"value"`
}

func main() {

	err := godotenv.Load("env/.env")
	if err != nil {
		log.Println("No .env file found or error loading .env")
	}

	influxURL := os.Getenv("INFLUXDB_URL")
	influxToken := os.Getenv("INFLUXDB_TOKEN")
	influxOrg := os.Getenv("INFLUXDB_ORG")
	influxBucket := os.Getenv("INFLUXDB_BUCKET")
	// robot := bot.NewRaspiRobot()
	// robot.Start()

	// create.L5XCreate()
	// db.Work()

	// config, err := email.LoadConfig()
	// if err != nil {
	// 	log.Fatal("Error loading config:", err)
	// }
	// recipient := "yourname@emailaddress.com"
	// subject := "This is an automated message from vtrgo."
	// message := "Please find your data report attached."
	// attachment := "output_files/Data_2024-06-10.xlsx"

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
	welcome("User")

	// Declare a PLC tag as a test integer variable
	myTag := db.PlcTag{
		Name: "TestDINT",
		Type: "dint",
	}

	myDintArray := db.PlcTag{
		Name:   "ArrayREAL",
		Type:   "[]real",
		Length: 10,
	}
	// Read PlcIpAddress and UseInflux from /env/vtrgo-config.json
	configFile, err := os.Open("env/vtrgo-config.json")
	if err != nil {
		log.Printf("Error opening config file: %v", err)
		return
	}
	defer configFile.Close()

	var config struct {
		PlcIpAddress    string `json:"PlcIpAddress"`
		ConfigUseInflux bool   `json:"UseInflux"`
	}
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Printf("Error decoding config file: %v", err)
		return
	}
	plcIpAddress := config.PlcIpAddress
	if plcIpAddress == "" {
		log.Printf("PlcIpAddress not found in config file")
		return
	}
	configUseInflux := config.ConfigUseInflux

	// Create a new PLC identity using the IP address of the Logix controller
	plc := plc.NewPLC(plcIpAddress)

	// Make a connection to the PLC
	err = plc.Connect()
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
	dataTagsDb, err = sql.Open("sqlite3", "./db_tags.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dataTagsDb.Close()

	// Initialize and create the database if it does not already exist
	err = db.InitTagDB(dataTagsDb, "dataTags")
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
	err = db.InitTagDB(alarmTagsDb, "alarmTags")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a connection to the plc_tags database
	usersDb, err = sql.Open("sqlite3", "./db_users.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer usersDb.Close()

	// Initialize and create the database if it does not already exist
	err = db.InitUserDB(usersDb)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

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
	})

	// Creates the /add-one endpoint to increment the test integer value and update it in the PLC
	http.HandleFunc("/add-one", func(write http.ResponseWriter, read *http.Request) {
		myTag.Value = myTag.Value.(int32) + 1
		plc.WriteTag(myTag.Name, myTag.Type, myTag.Value)
		log.Printf("Tag value updated: %v", myTag.Value)

		htmlResponse := fmt.Sprintf("<p>%s: <span id=\"tagValue\">%d</span></p>", myTag.Name, myTag.Value)
		write.Header().Set("Content-Type", "text/html")
		write.Write([]byte(htmlResponse))

	})

	// Creates the /update endpoint to increment the test integer value and store the new value in excel
	http.HandleFunc("/subtract-one", func(write http.ResponseWriter, read *http.Request) {
		myTag.Value = myTag.Value.(int32) - 1
		plc.WriteTag(myTag.Name, myTag.Type, myTag.Value)
		log.Printf("Tag value updated: %v", myTag.Value)

		htmlResponse := fmt.Sprintf("<p>%s: <span id=\"tagValue\">%d</span></p>", myTag.Name, myTag.Value)
		write.Header().Set("Content-Type", "text/html")
		write.Write([]byte(htmlResponse))
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-data-tags-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/data-tags-section.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-config-tags-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/config-section.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-alarm-tags-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/alarm-tags-section.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-users-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/users-section.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/load-metrics-section", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/metrics-section.html"))
		tmpl.Execute(w, nil)
	})

	// System data
	// date := time.Now().Format("2006-01-02")
	interval := 300 * time.Millisecond

	customerName := "VTR"
	recipeTag := "RecipeName"

	dataTriggerTag := "DataTrigger"
	dataResponseTag := "DataResponse"

	alarmTriggerTag := "AlarmTrigger"
	alarmResponseTag := "AlarmResponse"

	recipeName, err := plc.ReadTagString(recipeTag)
	if err != nil {
		log.Printf("Error reading tag: %v", err)
		return
	}

	dataFilePath := fmt.Sprintf("output_files/%s_%s-Data_%s.xlsx", customerName, recipeName, (time.Now().Format("2006-01-02")))
	alarmFilePath := fmt.Sprintf("output_files/%s_%s-Alarms_", customerName, recipeName)

	var influxDb *db.InfluxDBClient
	if configUseInflux {
		influxDb = db.NewInfluxDBClient(influxURL, influxToken, influxOrg, influxBucket)
	} else {
		influxDb = nil
	}
	log.Printf("InfluxDB URL: %s", influxURL)
	log.Printf("configUseInflux: %v", configUseInflux)
	dataTriggerChecker(dataTagsDb, plc, dataTriggerTag, dataResponseTag, dataFilePath, interval, influxDb, configUseInflux)
	alarmTriggerRoutine(alarmTagsDb, plc, alarmTriggerTag, alarmResponseTag, alarmFilePath, interval)
	// Sets up endpoint handlers for each function call
	// http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/add-tag", addDataTagHandler)
	http.HandleFunc("/remove-tag", removeDataTagHandler)
	http.HandleFunc("/list-tags", listDataTagsHandler)
	http.HandleFunc("/add-alarm-tag", addAlarmTagHandler)
	http.HandleFunc("/remove-alarm-tag", removeAlarmTagHandler)
	http.HandleFunc("/list-alarm-tags", listAlarmTagsHandler)
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
	http.HandleFunc("/check-alarms", checkAlarmsHandler)
	http.HandleFunc("/load-check-alarms", loadCheckAlarmsHandler)
	http.HandleFunc("/js/metricsChart.js", metricsChartHandler)

	log.Println("Server started at :8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
	// create.L5XCreate()

}

func metricsChartHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "resources/js/metricsChart.js")
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

func loadCheckAlarmsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/check-alarms.html")
}

// Handles the /list-tags endpoint for displaying all the tags stored in the plc_tags database
func listDataTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := db.FetchTags(dataTagsDb, "dataTags")
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
func addDataTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("tag")
		tagType := r.FormValue("type")
		lengthText := r.FormValue("length")
		length, err := strconv.Atoi(lengthText)
		if tag == "" || tagType == "" {
			http.Error(w, "Tag name and type are required", http.StatusBadRequest)
			return
		}

		db.InsertTag(dataTagsDb, "dataTags", tag, tagType, length)
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
func removeDataTagHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		tag := r.FormValue("removed-tag")
		if tag == "" {
			http.Error(w, "Tag name required", http.StatusBadRequest)
			return
		}

		err := db.RemoveTag(dataTagsDb, "dataTags", tag)
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
	tags, err := db.FetchTags(alarmTagsDb, "alarmTags")
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

		db.InsertTag(alarmTagsDb, "alarmTags", tag, tagType, length)
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

		err := db.RemoveTag(alarmTagsDb, "alarmTags", tag)
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
	users, err := db.FetchUsers(usersDb, false)
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

		err := db.InsertUser(usersDb, user, email, active)
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

		err := db.RemoveUser(usersDb, user)
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

// Handles the /list-tags endpoint for displaying all the tags stored in the plc_tags database
func checkAlarmsHandler(w http.ResponseWriter, r *http.Request) {
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

	tags, err := db.CheckAlarms(alarmsDb)
	if err != nil {
		log.Printf("Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<ul>")
	for _, tag := range tags {
		fmt.Fprintf(w, "<li>%s (%s)[%s]</li>", tag.Tag, tag.Message, tag.Trigger)
	}
	fmt.Fprintf(w, "</ul>")
}
