package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/AllenDang/giu"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var targetURL = flag.String("target-url", "http://0.0.0.0:8080/metrics", "the target URL")
var scrapeInterval = flag.Duration("scrape-interval", 250*time.Millisecond, "the scrape interval")
var metricsAddr = flag.String("metrics-addr", ":8080", "the UI metrics addr")

func main() {
	// parse flags
	flag.Parse()

	// run prometheus
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		panic(http.ListenAndServe(*metricsAddr, nil))
	}()

	// create window
	win := giu.NewMasterWindow("Promview", 1400, 900, 0)

	// run scraper
	go func() {
		for {
			// scrape metric families
			families, err := scrape(*targetURL)
			if err != nil {
				panic(err)
			}

			// ingest metric families
			err = ingest(families)
			if err != nil {
				panic(err)
			}

			// update
			giu.Update()

			// await next interval
			time.Sleep(*scrapeInterval)
		}
	}()

	// run ui code
	win.Run(func() {
		w, h := win.GetSize()
		giu.Window("Data").Pos(0, 0).Size(float32(w), float32(h)).Flags(giu.WindowFlagsNoResize | giu.WindowFlagsNoMove).Layout(
			giu.Custom(func() {
				// walk series
				walk(func(s *series) {
					// get data
					slice := s.slice()
					min, max := s.minMax()

					// make plot
					giu.Plot(s.name).AxisLimits(0, float64(len(slice)), min-5, max+5, giu.ConditionAlways).Plots(
						giu.PlotLine("", slice),
					).Build()
				})
			}),
		)
	})
}
