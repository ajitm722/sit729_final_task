package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"bytes"
	"io/ioutil"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	dt               = 0.04
	densityAir       = 1000.0
	Kp1              = 1000.0
	volO1i           = 30.0
	serverAddr       = "tcp://localhost:1883"                                                         // MQTT broker address
	displayServerURL = "http://budgethomedisplaylb-317811151.us-east-1.elb.amazonaws.com:3000/upload" // URL of the display server
)

var (
	TemperatureRoom1 = volO1i
)

// CreateCSVFile initializes the CSV file and returns the file and writer
func CreateCSVFile(filename string) (*os.File, *csv.Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, nil, err
	}

	writer := csv.NewWriter(file)

	// Write CSV header
	writer.Write([]string{
		"Time", "Reference Temperature", "Actual Temperature", "Error", "Control Input", "People In Room", "Humidity", "Annotation",
	})
	writer.Flush()

	return file, writer, nil
}

// handleMessage handles incoming MQTT messages and processes the data
func handleMessage(msg mqtt.Message, writer *csv.Writer) {
	// Parse the message payload
	payload := string(msg.Payload())
	fields := parsePayload(payload)
	if len(fields) < 4 {
		fmt.Println("Invalid payload:", payload)
		return
	}

	time, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}
	volR1, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		fmt.Println("Error parsing reference Temperature:", err)
		return
	}
	humidity, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		fmt.Println("Error parsing Humidity:", err)
		return
	}
	peopleInRoom, err := strconv.Atoi(fields[3])
	if err != nil {
		fmt.Println("Error parsing people in room:", err)
		return
	}

	var error1, mDot1 float64
	var annotation string

	if peopleInRoom > 0 {
		// Compute error
		error1 = volR1 - TemperatureRoom1

		// Compute control input
		mDot1 = Kp1 * error1

		// Compute true room Temperature
		TemperatureRoom1 += (mDot1 / densityAir) * dt
	} else {
		annotation = "No people in room"
	}

	// Write data to CSV
	writer.Write([]string{
		fmt.Sprintf("%.2f", time),
		fmt.Sprintf("%.2f", volR1),
		fmt.Sprintf("%.2f", TemperatureRoom1),
		fmt.Sprintf("%.2f", error1),
		fmt.Sprintf("%.2f", mDot1),
		fmt.Sprintf("%d", peopleInRoom),
		fmt.Sprintf("%.2f", humidity),
		annotation,
	})
	writer.Flush() // Ensure data is written to file
}

// parsePayload splits the incoming MQTT message payload into fields
func parsePayload(payload string) []string {
	return strings.Split(payload, ",")
}

// sendCSVData periodically sends the CSV file to the display server
func sendCSVData(filename string) {
	for {
		time.Sleep(13 * time.Second)

		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Failed to open CSV file: %v", err)
			continue
		}

		data, err := ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			log.Printf("Failed to read CSV file: %v", err)
			continue
		}

		resp, err := http.Post(displayServerURL, "text/csv", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Failed to send CSV data: %v", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to send CSV data, server responded with status: %d", resp.StatusCode)
		}
	}
}

func main() {
	// Set up the CSV file outside of the message handler
	file, writer, err := CreateCSVFile("temperature_data.csv")
	if err != nil {
		log.Fatalf("Could not open CSV file: %v", err)
	}
	defer file.Close() // Close file when the main function ends
	defer writer.Flush()

	// Set up MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(serverAddr)
	opts.SetClientID("serverID")
	opts.SetCleanSession(true)
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("Connected to MQTT broker")
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Println("Connection lost:", err)
	}

	// Create and start a new MQTT client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Subscribe to the topic and pass the writer by reference
	topic := "temperature_data"
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		handleMessage(msg, writer)
	})

	// Start the CSV sending thread
	go sendCSVData("temperature_data.csv")

	select {} // Keep the server running
}
