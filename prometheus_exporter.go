package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

var (
	invalidChars = regexp.MustCompile("[^a-zA-Z0-9_]")
	lastPush     = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jtimon_last_push_timestamp_seconds",
			Help: "Unix timestamp of the last received jtimon metrics push in seconds.",
		},
	)
)

type jtimonSample struct {
	ID        string
	Name      string
	Labels    map[string]string
	Value     float64
	Timestamp time.Time
}

type jtimonCollector struct {
	samples map[string]*jtimonSample
	mu      sync.Mutex
	ch      chan *jtimonSample
}

func newJTIMONCollector() *jtimonCollector {
	c := &jtimonCollector{
		ch:      make(chan *jtimonSample),
		samples: map[string]*jtimonSample{},
	}
	return c
}

func (c *jtimonCollector) getSamples() {
	ticker := time.NewTicker(time.Minute).C
	for {
		select {
		case s := <-c.ch:
			c.mu.Lock()
			c.samples[s.ID] = s
			//fmt.Println("One Sample Start")
			//fmt.Println(s.ID)
			//fmt.Println(s.Name)
			// fmt.Println(s.Labels)
			// fmt.Println(s.Value)
			// fmt.Println(s.Timestamp)
			// fmt.Println("One Sample End")
			c.mu.Unlock()

		case <-ticker:
			ageLimit := time.Now().Add(-time.Minute)
			c.mu.Lock()
			for k, sample := range c.samples {
				if ageLimit.After(sample.Timestamp) {
					delete(c.samples, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector
func (c *jtimonCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- lastPush

	c.mu.Lock()
	samples := make([]*jtimonSample, 0, len(c.samples))
	for _, sample := range c.samples {
		samples = append(samples, sample)
	}
	c.mu.Unlock()

	ageLimit := time.Now().Add(-time.Minute)
	for _, sample := range samples {
		if ageLimit.After(sample.Timestamp) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(sample.Name, "JTIMON Metric", []string{}, sample.Labels),
			prometheus.UntypedValue,
			sample.Value,
		)
	}
}

// Describe implements prometheus.Describe
func (c *jtimonCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastPush.Desc()
}

func addPrometheus(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	lastPush.Set(float64(time.Now().UnixNano()) / 1e9)
	c := jctx.promCollector
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
		tags["sensor"] = ocData.Path

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
		default:
			continue
		}

		fmt.Println("field:", field, "value:", fieldValue, "tags:", tags)

		sample := &jtimonSample{
			Name:      invalidChars.ReplaceAllString(field, "_"),
			Timestamp: time.Now(),
			Value:     fieldValue,
			Labels:    map[string]string{},
		}
		for k, v := range tags {
			sample.Labels[invalidChars.ReplaceAllString(k, "_")] = v
		}

		// Calculate a consistent unique ID for the sample.
		labelnames := make([]string, 0, len(sample.Labels))
		for k := range sample.Labels {
			labelnames = append(labelnames, k)
		}
		sort.Strings(labelnames)
		parts := make([]string, 0, len(sample.Labels)*2+1)
		parts = append(parts, invalidChars.ReplaceAllString(field, "_"))
		for _, l := range labelnames {
			parts = append(parts, l, sample.Labels[l])
		}
		sample.ID = fmt.Sprintf("%q", parts)

		c.ch <- sample
	}
}

func promInit() *jtimonCollector {

	prometheus.MustRegister(version.NewCollector("jtimon_exporter"))
	c := newJTIMONCollector()
	prometheus.MustRegister(c)

	go func() {
		go c.getSamples()
		addr := fmt.Sprintf("localhost:%d", *promPort)
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println(http.ListenAndServe(addr, nil))
	}()

	return c
}
