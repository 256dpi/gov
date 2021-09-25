package main

import (
	"math"
	"sync"

	dto "github.com/prometheus/client_model/go"
)

var mutex sync.RWMutex
var data = map[string]*series{}
var names []string

type kind int

const (
	gauge kind = iota
	counter
)

const storage = 100

type series struct {
	kind   kind
	name   string
	help   string
	list   [storage * 2]float64
	last   int
	length int
}

func (s *series) slice() []float64 {
	return s.list[s.last : s.last+s.length]
}

func (s *series) minMax() (float64, float64) {
	// find minimum and maximum
	slice := s.slice()
	min, max := slice[0], slice[0]
	for _, value := range slice {
		min = math.Min(min, value)
		max = math.Max(max, value)
	}

	return min, max
}

func ingest(families []dto.MetricFamily) error {
	// acquire mutex
	mutex.Lock()
	defer mutex.Unlock()

	// ingest metrics
	for _, family := range families {
		for _, metric := range family.Metric {
			err := ingestMetric(&family, metric)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ingestMetric(family *dto.MetricFamily, metric *dto.Metric) error {
	// get kind
	var knd kind
	switch *family.Type {
	case dto.MetricType_COUNTER:
		knd = counter
	case dto.MetricType_GAUGE:
		knd = gauge
	case dto.MetricType_SUMMARY, dto.MetricType_UNTYPED, dto.MetricType_HISTOGRAM:
		return nil
	}

	// get series
	srs, ok := data[*family.Name]
	if !ok {
		srs = &series{
			kind: knd,
			name: *family.Name,
			help: *family.Help,
		}
		data[*family.Name] = srs
		names = append(names, *family.Name)
	}

	// get value
	var value float64
	switch knd {
	case gauge:
		value = *metric.Gauge.Value
	case counter:
		value = *metric.Counter.Value
	}

	// increment position
	srs.last++
	if srs.last >= storage {
		srs.last = 0
	}

	// write values
	srs.list[srs.last] = value
	srs.list[storage+srs.last] = value

	// increment length
	if srs.length < storage {
		srs.length++
	}

	// metric.Label
	// metric.Gauge.Value
	// metric.Counter.Value
	// metric.Counter.Exemplar
	// metric.Summary.SampleSum
	// metric.Summary.SampleSum
	// metric.Summary.Quantile
	// metric.Untyped.Value
	// metric.Histogram.SampleCount
	// metric.Histogram.SampleSum
	// metric.Histogram.Bucket
	// metric.TimestampMs

	return nil
}

func walk(fn func(*series)) {
	// acquire mutex
	mutex.RLock()
	defer mutex.RUnlock()

	// yield series
	for _, name := range names {
		fn(data[name])
	}
}
