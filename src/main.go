package main

import "os"

var mqttURL string
var mqttUser string
var mqttPassword string
var prometheusURL string
var prometheusPath string

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
	prometheusURL = os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "0.0.0.0:2121"
	}
	prometheusPath = os.Getenv("PROMETHEUS_PATH")
	if prometheusPath == "" {
		prometheusPath = "/metrics"
	}
	initMessages()
	mqttRun()
	prometheusRun()
}
