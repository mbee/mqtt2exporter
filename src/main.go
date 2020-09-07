package main

import "os"

var mqttURL string
var mqttUser string
var mqttPassword string

func main() {
	mqttURL = os.Getenv("MQTT_URL")
	if mqttURL == "" {
		panic("MQTT_URL environment variable must be set")
	}
	mqttUser = os.Getenv("MQTT_USER")
	if mqttUser == "" {
		panic("MQTT_USER environment variable must be set")
	}
	mqttPassword = os.Getenv("MQTT_PASSWORD")
	if mqttPassword == "" {
		panic("MQTT_PASSWORD environment variable must be set")
	}
	initDevices()
	mqttRun()
	prometheusRun()
}
