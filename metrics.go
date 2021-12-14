package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var metricsMutex sync.RWMutex
var metricsTree = metricsNode{name: "root"}

type kind int

const (
	gauge kind = iota
	counter
)

type metricSeries struct {
	kind  kind
	name  string
	help  string
	dims  []string
	lists map[string]*list
}

func scrapeMetrics(url string, splitDepth int) error {
	// get families
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	// ensure close
	defer res.Body.Close()

	// determine format
	format := expfmt.ResponseFormat(res.Header)

	// create decoder
	dec := expfmt.NewDecoder(res.Body, format)

	// decode families
	var families []dto.MetricFamily
	for {
		var family dto.MetricFamily
		err = dec.Decode(&family)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		families = append(families, family)
	}

	// acquire mutex
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	// ingest metrics
	for _, family := range families {
		for _, metric := range family.Metric {
			err := ingestMetric(&family, metric, splitDepth)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ingestMetric(family *dto.MetricFamily, metric *dto.Metric, splitDepth int) error {
	// get kind
	var knd kind
	switch family.GetType() {
	case dto.MetricType_COUNTER:
		knd = counter
	case dto.MetricType_GAUGE, dto.MetricType_UNTYPED:
		knd = gauge
	case dto.MetricType_SUMMARY, dto.MetricType_HISTOGRAM:
		return nil
	}

	// check name
	if family.Name == nil {
		return fmt.Errorf("missing name")
	}

	// ensure node
	node := metricsTree.ensure(strings.SplitN(*family.Name, "_", splitDepth))

	// ensure series
	if node.series == nil {
		node.series = &metricSeries{
			kind:  knd,
			name:  *family.Name,
			help:  family.GetHelp(),
			lists: map[string]*list{},
		}
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
	list, ok := node.series.lists[dim]
	if !ok {
		list = newList(*seriesLength)
		node.series.lists[dim] = list
		node.series.dims = append(node.series.dims, dim)
	}

	// add value
	list.add(value, knd == counter)

	// TODO: Implement:
	//  metric.Summary.SampleSum
	//  metric.Summary.SampleSum
	//  metric.Summary.Quantile
	//  metric.Histogram.SampleCount
	//  metric.Histogram.SampleSum
	//  metric.Histogram.Bucket
	//  metric.TimestampMs

	return nil
}

func walkMetrics(node *metricsNode, fn func(*metricSeries)) {
	// acquire mutex
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	// yield series
	node.walk(func(m *metricsNode) {
		if m.series != nil {
			fn(m.series)
		}
	})
}

func withMetricsTree(fn func(node2 *metricsNode)) {
	// acquire mutex
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	// yield
	fn(&metricsTree)
}
