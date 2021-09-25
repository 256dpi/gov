package main

import (
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

type series struct {
	kind kind
	name string
	help string
	list *list
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
			list: newList(),
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

	// add value
	srs.list.add(value)

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
