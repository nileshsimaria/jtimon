package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// Prometheus does not like special characters, handle them.
func promName(input string) string {
	return promNameRegex.ReplaceAllString(input, "_")
}

type jtimonMetric struct {
	mapKey           string
	metricName       string
	metricLabels     map[string]string
	metricValue      float64
	metricExpiration time.Time
}

type jtimonPExporter struct {
	m  map[string]*jtimonMetric
	mu sync.Mutex
	ch chan *jtimonMetric
}

func newJTIMONPExporter() *jtimonPExporter {
	return (&jtimonPExporter{
		ch: make(chan *jtimonMetric),
		m:  map[string]*jtimonMetric{},
	})
}

func (c *jtimonPExporter) processJTIMONMetric() {
	ticker := time.NewTicker(time.Minute).C
	for {
		select {
		case s := <-c.ch:
			c.mu.Lock()
			c.m[s.mapKey] = s
			c.mu.Unlock()

		case <-ticker:
			ageLimit := time.Now().Add(-time.Minute)
			c.mu.Lock()
			for k, metric := range c.m {
				if ageLimit.After(metric.metricExpiration) {
					delete(c.m, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector
func (c *jtimonPExporter) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	metrics := make([]*jtimonMetric, 0, len(c.m))
	for _, metric := range c.m {
		metrics = append(metrics, metric)
	}
	c.mu.Unlock()

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(metric.metricName, "JTIMON Metric", []string{}, metric.metricLabels),
			prometheus.UntypedValue,
			metric.metricValue,
		)
	}
}

// Describe implements prometheus.Describe
func (c *jtimonPExporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func getMapKey(metric *jtimonMetric) string {
	labels := make([]string, 0, len(metric.metricLabels))

	for k := range metric.metricLabels {
		labels = append(labels, k)
	}

	sort.Strings(labels)

	mapKey := make([]string, 0, len(metric.metricLabels)*2+1)
	mapKey = append(mapKey, metric.metricName)

	for _, l := range labels {
		mapKey = append(mapKey, l, metric.metricLabels[l])
	}

	return fmt.Sprintf("%q", mapKey)
}

func addPrometheus(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	exporter := jctx.pExporter
	cfg := jctx.config

	prefix := ""

	for _, v := range ocData.Kv {
		switch {
		case v.Key == "__prefix__":
			prefix = v.GetStrValue()
			continue
		case strings.HasPrefix(v.Key, "__"):
			continue
		}

		key := v.Key
		if !strings.HasPrefix(key, "/") {
			key = prefix + v.Key
		}

		field, tags := spitTagsNPath(jctx, key)
		tags["device"] = cfg.Host

		var fieldValue float64

		switch v.Value.(type) {
		case *na_pb.KeyValue_DoubleValue:
			fieldValue = v.GetDoubleValue()
		case *na_pb.KeyValue_IntValue:
			fieldValue = float64(v.GetIntValue())
		case *na_pb.KeyValue_UintValue:
			fieldValue = float64(v.GetUintValue())
		case *na_pb.KeyValue_SintValue:
			fieldValue = float64(v.GetSintValue())
		case *na_pb.KeyValue_BoolValue:
			boolValue := v.GetBoolValue()
			if boolValue {
				fieldValue = 1
			} else {
				fieldValue = 0
			}
		case *na_pb.KeyValue_StrValue:
			floatVal, err := strconv.ParseFloat(v.GetStrValue(), 64)
			if err != nil {
				continue
			}
			fieldValue = floatVal
		default:
			continue
		}

		metric := &jtimonMetric{
			metricName:       promName(field),
			metricExpiration: time.Now(),
			metricValue:      fieldValue,
			metricLabels:     map[string]string{},
		}
		for k, v := range tags {
			metric.metricLabels[promName(k)] = v
		}

		metric.mapKey = getMapKey(metric)
		exporter.ch <- metric
	}
}

func promInit() *jtimonPExporter {

	c := newJTIMONPExporter()
	prometheus.MustRegister(c)

	go func() {
		go c.processJTIMONMetric()

		addr := fmt.Sprintf("%s:%d", *promHost, *promPort)
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println(http.ListenAndServe(addr, nil))
	}()

	return c
}
