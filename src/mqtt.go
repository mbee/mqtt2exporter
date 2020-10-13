package main

import (
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	metrics := getMetrics(msg.Topic(), string(msg.Payload()))
	if len(metrics) == 0 {
		return
	}
	exposeMetrics(metrics)
}

func mqttRun(mqttURL, mqttUser, mqttPassword string) {
	opts := mqtt.NewClientOptions().AddBroker(mqttURL).SetClientID("mqtt2prometheus")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPassword)
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("#", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}
