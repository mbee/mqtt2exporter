package lib

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goiiot/libmqtt"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Println(msg.Topic(), string(msg.Payload()))
	metrics := getMetrics(msg.Topic(), string(msg.Payload()))
	if len(metrics) == 0 {
		return
	}
	addSynonyms(metrics)
	exposeMetrics(metrics)
}

func MqttRun(mqttURL, mqttUser, mqttPassword, mqttClientID string) {
	client, err := libmqtt.NewClient(
		libmqtt.WithKeepalive(10, 1.2),
		libmqtt.WithAutoReconnect(true),
		libmqtt.WithBackoffStrategy(time.Second, 5*time.Second, 1.2),
		libmqtt.WithRouter(libmqtt.NewRegexRouter()),
	)
	if err != nil {
		panic("create mqtt client failed")
	}
	client.ConnectServer(mqttURL,
		libmqtt.WithCustomTLS(nil),
		libmqtt.WithClientID(mqttClientID),
		libmqtt.WithIdentity(mqttUser, mqttPassword),
		libmqtt.WithConnHandleFunc(func(client libmqtt.Client, server string, code byte, err error) {
			if err != nil {
				panic(err)
			}
			if code != libmqtt.CodeSuccess {
				panic(code)
			}
			log.Printf("Connected to %s with user %s", mqttURL, mqttUser)
			client.HandleTopic(".*", func(client libmqtt.Client, topic string, qos libmqtt.QosLevel, msg []byte) {
				log.Println(topic, string(msg))
			})
		}),
	)
}
