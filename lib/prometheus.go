package lib

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricType struct {
	name   string
	value  string
	labels map[string]string
}

type gauge struct {
	gaugeVec *prometheus.GaugeVec
	labels   []string
}

var gaugeVecMap = map[string]gauge{}

func testSliceOfStrings(a, b []string) {}

func getGaugeVec(name string, labelNames []string) (*gauge, error) {
	// TODO => check if the labels are the same
	if gaugeVec, found := gaugeVecMap[name]; found {
		a := gaugeVec.labels
		b := labelNames
		sort.Strings(a)
		sort.Strings(b)
		if strings.Join(a, ":") != strings.Join(b, ":") {
			return nil, fmt.Errorf("Same gauge name, different labels : %s => [%s] != [%s]", name, a, b)
		}
		return &gaugeVec, nil
	}
	log.Printf("creating new gauge for %s [%s]\n", name, labelNames)
	gauge := gauge{
		labels: labelNames,
		gaugeVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
		}, labelNames),
	}

	gaugeVecMap[name] = gauge
	err := prometheus.Register(gauge.gaugeVec)
	if err != nil {
		return nil, err
	}
	return &gauge, nil
}

func exposeMetrics(metrics []metricType) {
	for _, metric := range metrics {
		labels := make([]string, 0, len(metric.labels))
		for k := range metric.labels {
			labels = append(labels, k)
		}
		gauge, err := getGaugeVec(metric.name, labels)
		if err != nil {
			log.Println(err)
			continue
		}
		value, err := strconv.ParseFloat(metric.value, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		labelValues := make([]string, 0, len(gauge.labels))
		for _, label := range gauge.labels {
			labelValues = append(labelValues, metric.labels[label])
		}
		gauge.gaugeVec.WithLabelValues(labelValues...).Set(value)
		gaugeTimestamp, err := getGaugeVec(metric.name+"_last_seconds", labels)
		if err != nil {
			log.Println(err)
			continue
		}
		now := time.Now()
		gaugeTimestamp.gaugeVec.WithLabelValues(labelValues...).Set((time.Duration(now.UnixNano()).Seconds()))
	}
}

func PrometheusRun(prometheusURL, prometheusPath string) {
	http.Handle(prometheusPath, promhttp.Handler())
	log.Printf("Listening to http://%s%s", prometheusURL, prometheusPath)
	http.ListenAndServe(prometheusURL, nil)
}
