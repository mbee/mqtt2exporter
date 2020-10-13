package main

import "os"

func main() {
	mqttURL := os.Getenv("MQTT_URL")
	if mqttURL == "" {
		panic("MQTT_URL environment variable must be set")
	}
	mqttUser := os.Getenv("MQTT_USER")
	if mqttUser == "" {
		panic("MQTT_USER environment variable must be set")
	}
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	if mqttPassword == "" {
		panic("MQTT_PASSWORD environment variable must be set")
	}
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "0.0.0.0:2121"
	}
	prometheusPath := os.Getenv("PROMETHEUS_PATH")
	if prometheusPath == "" {
		prometheusPath = "/metrics"
	}
	devicesFilePath := os.Getenv("DEVICE_PATH")
	if devicesFilePath == "" {
		devicesFilePath = "static/messages"
	}
	initMessages(devicesFilePath)
	mqttRun(mqttURL, mqttUser, mqttPassword)
	prometheusRun(prometheusURL, prometheusPath)
}
