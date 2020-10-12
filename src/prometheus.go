package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricType struct {
	name        string
	value       string
	labels      []string
	labelValues []string
}

var gaugeVecMap = map[string]*prometheus.GaugeVec{}

func getGaugeVec(name string, labelNames []string) *prometheus.GaugeVec {
	// TODO => check if the labels are the same
	if gaugeVec, found := gaugeVecMap[name]; found {
		return gaugeVec
	}
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
	}, labelNames)
	gaugeVecMap[name] = gaugeVec
	err := prometheus.Register(gaugeVec)
	if err != nil {
		log.Println(err)
	}
	return gaugeVec
}

func exposeMetrics(metrics []metricType) {
	for _, metric := range metrics {
		gauge := getGaugeVec(metric.name, metric.labels)
		value, err := strconv.ParseFloat(metric.value, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		gauge.WithLabelValues(metric.labelValues...).Set(value)
	}
}

func prometheusRun() {
	http.Handle(prometheusPath, promhttp.Handler())
	http.ListenAndServe(prometheusURL, nil)
}
