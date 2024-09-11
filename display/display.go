package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	csvData         string
	mongoClient     *mongo.Client
	mongoCollection *mongo.Collection
)

// handleUpload handles the incoming CSV data
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	csvData = string(data)
	fmt.Fprintln(w, "CSV data received successfully")
}

// handleDisplay displays the latest temperature, humidity, and people in room data
func handleDisplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var latestTemperature, latestHumidity, latestPeopleInRoom string

	// Split CSV data into lines
	lines := strings.Split(csvData, "\n")
	if len(lines) > 1 {
		// Get the last line of the CSV (latest data)
		lastLine := lines[len(lines)-2] // Last line of actual data (ignoring the header)

		// Split the last line into fields
		fields := strings.Split(lastLine, ",")
		if len(fields) >= 7 {
			latestTemperature = fields[2]  // Actual Temperature
			latestHumidity = fields[6]     // Humidity
			latestPeopleInRoom = fields[5] // People In Room
		}
	}

	// Render the HTML with the latest data
	fmt.Fprintf(w, "<html><body><h1 style='font-size:48px;'>Budget Home</h1><p><strong>Latest Temperature:</strong> %s</p><p><strong>Latest Humidity:</strong> %s</p><p><strong>People In Room:</strong> %s</p></body></html>",
		latestTemperature, latestHumidity, latestPeopleInRoom)
}

// saveToMongoDB saves the latest data to MongoDB Atlas
func saveToMongoDB(temperature, humidity, peopleInRoom string) error {
	collection := mongoCollection

	// Create a document to insert
	document := bson.D{
		{Key: "timestamp", Value: time.Now()},
		{Key: "temperature", Value: temperature},
		{Key: "humidity", Value: humidity},
		{Key: "people_in_room", Value: peopleInRoom},
	}

	// Print debug information
	fmt.Printf("Persisting data to MongoDB: %+v\n", document)

	_, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		fmt.Printf("Error persisting data to MongoDB: %v\n", err)
	}
	return err
}

// periodicallySaveToMongoDB triggers the save function every 5 seconds
func periodicallySaveToMongoDB() {
	for {
		if csvData != "" {
			// Extract the latest data from CSV and save to MongoDB
			var latestTemperature, latestHumidity, latestPeopleInRoom string

			lines := strings.Split(csvData, "\n")
			if len(lines) > 1 {
				lastLine := lines[len(lines)-2]
				fields := strings.Split(lastLine, ",")
				if len(fields) >= 7 {
					latestTemperature = fields[2]
					latestHumidity = fields[6]
					latestPeopleInRoom = fields[5]
				}
			}

			if latestTemperature != "" && latestHumidity != "" && latestPeopleInRoom != "" {
				saveToMongoDB(latestTemperature, latestHumidity, latestPeopleInRoom)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	// MongoDB connection setup
	clientOptions := options.Client().ApplyURI("mongodb+srv://ajitmudgerikar:ajitsit729@ajitsit729.p7gfv.mongodb.net/?retryWrites=true&w=majority&appName=AjitSIT729")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	mongoClient = client
	mongoCollection = client.Database("Budget_Home").Collection("Latest_Sensor_Data")

	// Start the periodic MongoDB saving in a separate goroutine
	go periodicallySaveToMongoDB()

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/display", handleDisplay)

	fmt.Println("Display server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
