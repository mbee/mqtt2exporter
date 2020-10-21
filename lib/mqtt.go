package lib

import (
	"log"
	"time"

	"github.com/goiiot/libmqtt"
)

// MqttRun will run the mqtt listener
func MqttRun(mqttURL, mqttUser, mqttPassword, mqttClientID string) {
	client, err := libmqtt.NewClient(
		libmqtt.WithKeepalive(10, 1.2),
		libmqtt.WithAutoReconnect(true),
		libmqtt.WithBackoffStrategy(time.Second, 5*time.Second, 1.2),
		libmqtt.WithRouter(libmqtt.NewRegexRouter()),
		libmqtt.WithCleanSession(true),
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
				metrics := getMetrics(topic, string(msg))
				if len(metrics) == 0 {
					return
				}
				addSynonyms(metrics)
				exposeMetrics(metrics)
			})
			client.Subscribe([]*libmqtt.Topic{
				{Name: "#"},
			}...)
		}),
	)
}
