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
	mqttClientID := os.Getenv("MQTT_CLIENT_ID")
	if mqttClientID == "" {
		mqttClientID = "mqtt2exporter"
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
		devicesFilePath = "static/devices/"
	}
	synonymsFilePath := os.Getenv("SYNONYM_PATH")
	if synonymsFilePath == "" {
		synonymsFilePath = "static/synonyms/"
	}
	initMessages(devicesFilePath)
	initSynonyms(synonymsFilePath)
	mqttRun(mqttURL, mqttUser, mqttPassword, mqttClientID)
	prometheusRun(prometheusURL, prometheusPath)
}
