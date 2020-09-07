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

type file struct {
	Devices []Device `yaml:"devices"`
}
type Labels struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
type Metric struct {
	Name   string   `yaml:"name"`
	Value  string   `yaml:"value"`
	Labels []Labels `yaml:"labels"`
}
type Device struct {
	Name            string   `yaml:"name"`
	MessageType     string   `yaml:"message_type"`
	TopicRe         string   `yaml:"topic_re"`
	Topic           string   `yaml:"topic"`
	Metric          []Metric `yaml:"metric"`
	topicCompiledRe *regexp.Regexp
}

type regexpDenorm map[*regexp.Regexp]Device
type noregexpDenorm map[string]Device

var devices file
var regexps = regexpDenorm{}
var noregexps = noregexpDenorm{}

func addYaml(content []byte) {
	err := yaml.Unmarshal(content, &devices)
	if err != nil {
		panic(err)
	}
	for _, device := range devices.Devices {
		if device.Topic != "" {
			noregexps[device.Topic] = device
			continue
		}
		device.topicCompiledRe, err = regexp.Compile(device.TopicRe)
		if err != nil {
			log.Printf("Unable to parse regex '%s' for device '%s'", device.TopicRe, device.Name)
			continue
		}
		regexps[device.topicCompiledRe] = device
	}
}

func initDevices() {
	err := filepath.Walk("static/devices", func(path string, info os.FileInfo, err error) error {
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

func getMetrics(path string, message string) []metricType {
	result := []metricType{}

	// let's take care of topic with no regexp first
	for topic, device := range noregexps {
		if topic == path {
			result = append(result, metricType{
				name:  device.Name,
				value: message,
			})
			return result
		}
	}

	// then, let's have a look to topic with regexp
	for re, device := range regexps {
		match := re.FindStringSubmatch(path)
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
		for _, metric := range device.Metric {
			name := replace(metric.Name, replacement)
			value := message
			if device.MessageType == "json" {
				value = getValueFromJSONMessage(message, metric.Value)
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
