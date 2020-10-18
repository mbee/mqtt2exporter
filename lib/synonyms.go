package lib

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SynonymsFile contains all synonyms coming from files
type SynonymsFile struct {
	Synonyms []Synonym `yaml:"synonyms"`
}

// Synonym contains the name of a device for a dedicated id and type
type Synonym struct {
	DeviceID   string `yaml:"device_id"`
	DeviceType string `yaml:"device_type"`
	DeviceName string `yaml:"device_name"`
}

// will contatain device_type:device_id together for the key
// and set device_name as value
var indexedSynonyms = map[string]string{}

func readSynonymFile(content []byte) {
	synonymsFils := SynonymsFile{}
	err := yaml.Unmarshal(content, &synonymsFils)
	if err != nil {
		panic(err)
	}
	for _, synonym := range synonymsFils.Synonyms {
		if synonym.DeviceID == "" || synonym.DeviceType == "" || synonym.DeviceName == "" {
			log.Printf("Error reading %s:%s:%s from synonymFile\n", synonym.DeviceID, synonym.DeviceType, synonym.DeviceName)
			continue
		}
		indexedSynonyms[synonym.DeviceType+":"+synonym.DeviceID] = synonym.DeviceName
	}
}

func InitSynonyms(synonymsFilePath string) {
	err := filepath.Walk(synonymsFilePath, func(path string, info os.FileInfo, err error) error {
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
		readSynonymFile(content)
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func addSynonyms(metrics []metricType) {
	for _, metric := range metrics {
		deviceID, okDeviceID := metric.labels["device_id"]
		deviceType, okDeviceType := metric.labels["device_type"]
		_, okDeviceName := metric.labels["device_name"]
		if !okDeviceID || !okDeviceType || okDeviceName {
			continue
		}
		key := deviceType + ":" + deviceID
		if name, found := indexedSynonyms[key]; found {
			metric.labels["device_name"] = name
		} else {
			metric.labels["device_name"] = metric.labels["device_id"]
		}
	}
}
