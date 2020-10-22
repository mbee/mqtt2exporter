package lib

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
	Count  bool    `yaml:"count"`
	Labels []Label `yaml:"labels"`
}

// Message decypher one or several mqtt messages
type Message struct {
	Type            string   `yaml:"type"`
	MessageType     string   `yaml:"message_type"`
	TopicRe         string   `yaml:"topic_re"`
	Topic           string   `yaml:"topic"`
	MetricName      string   `yaml:"metric_name"`
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

func readMessageFile(content []byte) {
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
			log.Printf("Unable to parse regexp '%s' for message '%s'", message.TopicRe, message.Type)
			continue
		}
		mqttPerRegexps[message.topicCompiledRe] = message
	}
}

// InitMessages will read the devices file
func InitMessages(devicesFilePath string) {
	err := filepath.Walk(devicesFilePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(err)
			return nil
		}
		log.Printf("Reading %s\n", path)
		readMessageFile(content)
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// replace takes a map of replacement string and will look for each of the keys, and replace them with the value.
// replacement['%toto%']
func replace(from string, replacement map[string]string) string {
	result := from
	// speed-up the process if there's nothing to replace
	if !strings.ContainsRune(from, '%') {
		return from
	}
	for key, value := range replacement {
		result = strings.ReplaceAll(result, key, value)
	}
	return result
}

func getValueFromJSONMessage(json, jqpath string) string {
	op, err := jq.Parse(jqpath)
	if err != nil {
		return ""
	}
	value, err := op.Apply([]byte(json))
	if err != nil {
		return ""
	}
	return string(value)
}

func getMetricsPerExactName(mqttTopic string, mqttMessage string) []metricStruct {
	result := []metricStruct{}
	// let's take care of topic with no regexp first
	for topic, message := range mqttPerName {
		if topic == mqttTopic {
			result = append(result, metricStruct{
				name:   message.MetricName,
				value:  mqttMessage,
				labels: map[string]string{"device_type": message.Type},
			})
			return result
		}
	}
	return result
}

func getMetricsPerRegexp(mqttTopic string, mqttMessage string) []metricStruct {
	result := []metricStruct{}
	for re, message := range mqttPerRegexps {
		match := re.FindStringSubmatch(mqttTopic)
		if match == nil {
			continue
		}

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
				if value == "" {
					continue
				}
			}
			labels := map[string]string{"device_type": message.Type}
			for _, label := range metric.Labels {
				labels[replace(label.Name, replacement)] = replace(label.Value, replacement)
			}
			metricType := GAUGE
			if metric.Count {
				metricType = COUNTER
			}
			result = append(result, metricStruct{
				name:   name,
				value:  value,
				labels: labels,
				mType:  metricType,
			})
		}
	}
	return result
}

// getMetrics will compare mqttTopic with:
// - topics from yaml files with no regexp
// - topics from yaml files with regexp
// If it founds a topic with no regexp, it does not look for topics with regexp.
func getMetrics(mqttTopic string, mqttMessage string) []metricStruct {
	// let's take care of topic with no regexp first
	result := getMetricsPerExactName(mqttTopic, mqttMessage)
	if len(result) != 0 {
		return result
	}

	// then, let's have a look to topic with regexp
	return getMetricsPerRegexp(mqttTopic, mqttMessage)
}
