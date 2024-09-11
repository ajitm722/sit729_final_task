package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	serverAddr = "tcp://localhost:1883" // MQTT broker address
	tEnd       = 100.0
	dt         = 0.04
	volR1i     = 70.0
)

type DataPacket struct {
	Time         float64
	VolR1        float64
	Humidity     float64
	PeopleInRoom int
}

func main() {
	// Set up MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(serverAddr)
	opts.SetClientID("clientID")
	opts.SetCleanSession(true)

	// Create and start a new MQTT client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Error connecting to MQTT broker:", token.Error())
		return
	}
	defer client.Disconnect(250)

	steps := int(tEnd / dt)
	lastVolR1 := volR1i // Initialize with the initial Temperature reference

	for i := 1; i < steps; i++ {
		packet := DataPacket{
			Time: float64(i) * dt,
		}

		if i < 300 {
			packet.VolR1 = volR1i
			packet.PeopleInRoom = 3 // Random number between 0 and 5
			packet.Humidity = 30.0  // Placeholder value
		} else if i < 600 {
			packet.VolR1 = 20
			packet.PeopleInRoom = 2 // Random number between 0 and 5
			packet.Humidity = 40.0  // Placeholder value
		} else if i < 900 {
			packet.VolR1 = 90
			packet.PeopleInRoom = 0 // Random number between 0 and 5
			packet.Humidity = 10.0  // Placeholder value
		} else if i < 1200 {
			packet.VolR1 = 30
			packet.PeopleInRoom = 4 // Random number between 0 and 5
			packet.Humidity = 60.0  // Placeholder value
		} else if i < 1500 {
			packet.VolR1 = 80
			packet.PeopleInRoom = 1 // Random number between 0 and 5
			packet.Humidity = 80.0  // Placeholder value
		} else if i < 1800 {
			packet.VolR1 = 10
			packet.PeopleInRoom = 5 // Random number between 0 and 5
			packet.Humidity = 90.0  // Placeholder value
		} else if i < 2100 {
			packet.VolR1 = 95
			packet.PeopleInRoom = 0 // Random number between 0 and 5
			packet.Humidity = 20.0  // Placeholder value
		} else {
			packet.VolR1 = 50
			packet.PeopleInRoom = 2 // Random number between 0 and 5
			packet.Humidity = 50.0  // Placeholder value
		}

		// Publish the data packet to the MQTT broker
		topic := "temperature_data"
		payload := fmt.Sprintf("%f,%f,%f,%d", packet.Time, packet.VolR1, packet.Humidity, packet.PeopleInRoom)
		token := client.Publish(topic, 0, false, payload)
		token.Wait()

		// Print statement when reference Temperature changes
		if packet.VolR1 != lastVolR1 {
			fmt.Printf("Reference Temperature changed to: %.2f at time %.2f\n", packet.VolR1, packet.Time)
			lastVolR1 = packet.VolR1 // Update lastVolR1
		}

		// Wait before sending the next data packet
		time.Sleep(time.Duration(dt*1000) * time.Millisecond)
	}
}
