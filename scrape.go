package main

import (
	"io"
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func scrape(url string) ([]dto.MetricFamily, error) {
	// get families
	res, err := http.Get(url)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		families = append(families, family)
	}

	return families, nil
}
