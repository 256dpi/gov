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
	// check name
	if family.Name == nil {
		return fmt.Errorf("missing name")
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

	// add metric
	switch family.GetType() {
	case dto.MetricType_COUNTER:
		getList(*family.Name, family.GetHelp(), dim, splitDepth).addDiff(*metric.Counter.Value)
	case dto.MetricType_GAUGE:
		getList(*family.Name, family.GetHelp(), dim, splitDepth).add(*metric.Gauge.Value)
	case dto.MetricType_UNTYPED:
		getList(*family.Name, family.GetHelp(), dim, splitDepth).add(*metric.Gauge.Value)
	case dto.MetricType_SUMMARY:
		getList(*family.Name+":count", family.GetHelp(), dim, splitDepth).addDiff(float64(*metric.Summary.SampleCount))
		getList(*family.Name+":mean", family.GetHelp(), dim, splitDepth).addMean(*metric.Summary.SampleSum, float64(*metric.Summary.SampleCount))
		for _, bucket := range metric.Summary.Quantile {
			getList(*family.Name+":"+f2s(*bucket.Quantile), family.GetHelp(), dim, splitDepth).add(*bucket.Value)
		}
	case dto.MetricType_HISTOGRAM:
		getList(*family.Name+":count", family.GetHelp(), dim, splitDepth).addDiff(float64(*metric.Histogram.SampleCount))
		getList(*family.Name+":mean", family.GetHelp(), dim, splitDepth).addMean(*metric.Histogram.SampleSum, float64(*metric.Histogram.SampleCount))
		for i, bucket := range metric.Histogram.Bucket {
			count := *bucket.CumulativeCount
			if i > 0 {
				count -= *metric.Histogram.Bucket[i-1].CumulativeCount
			}
			getList(*family.Name+":"+f2s(*bucket.UpperBound), family.GetHelp(), dim, splitDepth).addDiff(float64(count))
		}
	}

	return nil
}

func getList(name, help, dim string, splitDepth int) *list {
	// ensure node
	node := metricsTree.ensure(strings.SplitN(name, "_", splitDepth))

	// ensure series
	if node.series == nil {
		node.series = &metricSeries{
			name:  name,
			help:  help,
			lists: map[string]*list{},
		}
	}

	// get list
	list, ok := node.series.lists[dim]
	if !ok {
		list = newList(*seriesLength)
		node.series.lists[dim] = list
		node.series.dims = append(node.series.dims, dim)
	}

	return list
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
