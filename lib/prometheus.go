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

var gaugeVecMap = map[string]*gauge{}
var registry = prometheus.NewRegistry()

func testSliceOfStrings(a, b []string) {}

func getGaugeVec(name string, labels map[string]string) (*prometheus.Gauge, error) {
	// TODO => check if the labels are the same
	labelNames := make([]string, 0, len(labels))
	for k := range labels {
		labelNames = append(labelNames, k)
	}
	sort.Strings(labelNames)
	if _, found := gaugeVecMap[name]; !found {
		gauge := gauge{
			labels: labelNames,
			gaugeVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: name,
			}, labelNames),
		}
		err := registry.Register(gauge.gaugeVec)
		if err != nil {
			return nil, err
		}
		gaugeVecMap[name] = &gauge
	} else {
		a := gaugeVecMap[name].labels
		sort.Strings(a)
		if strings.Join(a, ":") != strings.Join(labelNames, ":") {
			return nil, fmt.Errorf("Same gauge name, different labels : %s => [%s] != [%s]", name, a, labelNames)
		}
	}
	gaugeVec := gaugeVecMap[name].gaugeVec
	g, err := gaugeVec.GetMetricWith(labels)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func exposeMetrics(metrics []metricType) {
	for _, metric := range metrics {
		value, err := strconv.ParseFloat(metric.value, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		gauge, err := getGaugeVec(metric.name, metric.labels)
		if err != nil {
			log.Println(err)
			continue
		}
		(*gauge).Set(value)
		gauge, err = getGaugeVec(metric.name+"_last_seconds", metric.labels)
		if err != nil {
			log.Println(err)
			continue
		}
		now := time.Now()
		(*gauge).Set((time.Duration(now.UnixNano()).Seconds()))
	}
}

// PrometheusRun runs the http listener
func PrometheusRun(prometheusURL, prometheusPath string) {
	http.Handle(prometheusPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Printf("Listening to http://%s%s", prometheusURL, prometheusPath)
	http.ListenAndServe(prometheusURL, nil)
}
