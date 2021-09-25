package main

import (
	"strings"
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
	kind  kind
	name  string
	help  string
	dims  []string
	lists map[string]*list
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
	case dto.MetricType_GAUGE, dto.MetricType_UNTYPED:
		knd = gauge
	case dto.MetricType_SUMMARY, dto.MetricType_HISTOGRAM:
		return nil
	}

	// get series
	srs, ok := data[*family.Name]
	if !ok {
		srs = &series{
			kind:  knd,
			name:  *family.Name,
			help:  *family.Help,
			lists: map[string]*list{},
		}
		data[*family.Name] = srs
		names = append(names, *family.Name)
	}

	// get value
	var value float64
	switch knd {
	case gauge:
		if metric.Gauge != nil {
			value = *metric.Gauge.Value
		} else if metric.Untyped != nil {
			value = *metric.Untyped.Value
		}
	case counter:
		value = *metric.Counter.Value
	}

	// get dimension
	dim := "default"
	if len(metric.Label) > 0 {
		pairs := make([]string, 0, len(metric.Label))
		for _, label := range metric.Label {
			pairs = append(pairs, *label.Name+":"+*label.Value)
		}
		dim = strings.Join(pairs, " ")
	}

	// get list
	list, ok := srs.lists[dim]
	if !ok {
		list = newList()
		srs.lists[dim] = list
		srs.dims = append(srs.dims, dim)
	}

	// add value
	list.add(value, knd == counter)

	// metric.Summary.SampleSum
	// metric.Summary.SampleSum
	// metric.Summary.Quantile
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
