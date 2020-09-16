package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/savaki/jq"
	"gopkg.in/yaml.v3"
)

// YamlFiles contain all yaml files describing the mqtt messages to monitor
type YamlFiles struct {
	Messages []Message `yaml:"messages"`
}

// Label describes the prometheus labels which will appear on prometheus metrics
type Label struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// Metric describes how to build a prometheus metric from the mqtt
type Metric struct {
	Name   string  `yaml:"name"`
	Value  string  `yaml:"value"`
	Labels []Label `yaml:"labels"`
}

// Message decypher one or several mqtt messages
type Message struct {
	Name            string   `yaml:"name"`
	MessageType     string   `yaml:"message_type"`
	TopicRe         string   `yaml:"topic_re"`
	Topic           string   `yaml:"topic"`
	Metric          []Metric `yaml:"metric"`
	topicCompiledRe *regexp.Regexp
}

var yamlFiles *YamlFiles
var mqttPerRegexps = map[*regexp.Regexp]Message{}
var mqttPerName = map[string]Message{}

func getYamlFiles() *YamlFiles {
	if yamlFiles == nil {
		yamlFiles = &YamlFiles{}
	}
	return yamlFiles
}

func addYaml(content []byte) {
	err := yaml.Unmarshal(content, getYamlFiles())
	if err != nil {
		panic(err)
	}
	for _, message := range yamlFiles.Messages {
		if message.Topic != "" {
			mqttPerName[message.Topic] = message
			continue
		}
		message.topicCompiledRe, err = regexp.Compile(message.TopicRe)
		if err != nil {
			log.Printf("Unable to parse regex '%s' for message '%s'", message.TopicRe, message.Name)
			continue
		}
		mqttPerRegexps[message.topicCompiledRe] = message
	}
}

func initMessages() {
	err := filepath.Walk("static/messages", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(err)
			return nil
		}
		log.Printf("Reading %s\n", path)
		addYaml(content)
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// TODO => cache the result
func replace(from string, replacement map[string]string) string {
	result := from
	if !strings.ContainsRune(result, '%') {
		return result
	}
	for key, value := range replacement {
		result = strings.ReplaceAll(result, key, value)
	}
	return result
}

func getValueFromJSONMessage(json, jqpath string) string {
	op, err := jq.Parse(jqpath)
	if err != nil {
		log.Printf("unable to parse the jqpath <%s>: %v\n", jqpath, err)
		return "0"
	}
	value, err := op.Apply([]byte(json))
	if err != nil {
		log.Printf("unable to get the jq path <%s> from message <%s>: %v\n", jqpath, json, err)
		return "0"
	}
	return string(value)
}

func getMetrics(mqttTopic string, mqttMessage string) []metricType {
	result := []metricType{}

	// let's take care of topic with no regexp first
	for topic, message := range mqttPerName {
		if topic == mqttTopic {
			result = append(result, metricType{
				name:  message.Name,
				value: mqttMessage,
			})
			return result
		}
	}

	// then, let's have a look to topic with regexp
	for re, message := range mqttPerRegexps {
		match := re.FindStringSubmatch(mqttTopic)
		if match == nil {
			continue
		}
		// TODO ensure there is no "message" key
		replacement := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				replacement["%"+name+"%"] = match[i]
			}
		}
		for _, metric := range message.Metric {
			name := replace(metric.Name, replacement)
			value := mqttMessage
			if message.MessageType == "json" {
				value = getValueFromJSONMessage(mqttMessage, metric.Value)
			}
			labels := []string{}
			labelValues := []string{}
			for _, label := range metric.Labels {
				labels = append(labels, replace(label.Name, replacement))
				labelValues = append(labelValues, replace(label.Value, replacement))
			}
			result = append(result, metricType{
				name:        name,
				value:       value,
				labels:      labels,
				labelValues: labelValues,
			})
		}
	}
	return result
}
