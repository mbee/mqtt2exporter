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

type metricType int

const (
	GAUGE metricType = iota
	COUNTER
)

type metricStruct struct {
	name   string
	value  string
	labels map[string]string
	mType  metricType
}

type gauge struct {
	gaugeVec *prometheus.GaugeVec
	labels   []string
}

type counter struct {
	counterVec *prometheus.CounterVec
	labels     []string
}

var gaugeVecMap = map[string]*gauge{}
var counterVecMap = map[string]*counter{}
var registry = prometheus.NewRegistry()

func testSliceOfStrings(a, b []string) {}

func getGaugeVec(name string, labels map[string]string) (*prometheus.Gauge, error) {
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

func getCounterVec(name string, labels map[string]string) (*prometheus.Counter, error) {
	labelNames := make([]string, 0, len(labels))
	for k := range labels {
		labelNames = append(labelNames, k)
	}
	sort.Strings(labelNames)
	if _, found := counterVecMap[name]; !found {
		counter := counter{
			labels: labelNames,
			counterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: name,
			}, labelNames),
		}
		err := registry.Register(counter.counterVec)
		if err != nil {
			return nil, err
		}
		counterVecMap[name] = &counter
	} else {
		a := counterVecMap[name].labels
		sort.Strings(a)
		if strings.Join(a, ":") != strings.Join(labelNames, ":") {
			return nil, fmt.Errorf("Same counter name, different labels : %s => [%s] != [%s]", name, a, labelNames)
		}
	}
	counterVec := counterVecMap[name].counterVec
	c, err := counterVec.GetMetricWith(labels)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func sanitizeValue(val string) string {
	return strings.Trim(val, "\" \t")
}

func exposeMetrics(metrics []metricStruct) {
	for _, metric := range metrics {
		if metric.mType == GAUGE {
			value, err := strconv.ParseFloat(sanitizeValue(metric.value), 64)
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
		if metric.mType == COUNTER {
			labels := metric.labels
			labels["value"] = strings.Trim(metric.value, "\"")
			counter, err := getCounterVec(metric.name, metric.labels)
			if err != nil {
				log.Println(err)
				continue
			}
			(*counter).Inc()
			gauge, err := getGaugeVec(metric.name+"_last_seconds", metric.labels)
			if err != nil {
				log.Println(err)
				continue
			}
			now := time.Now()
			(*gauge).Set((time.Duration(now.UnixNano()).Seconds()))
		}
	}
}

// PrometheusRun runs the http listener
func PrometheusRun(prometheusURL, prometheusPath string) {
	http.Handle(prometheusPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Printf("Listening to http://%s%s", prometheusURL, prometheusPath)
	http.ListenAndServe(prometheusURL, nil)
}
