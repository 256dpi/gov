package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/AllenDang/giu"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var seriesLength = flag.Int("series-length", 100, "the series length")
var targetURL = flag.String("target-url", "http://0.0.0.0:8080/metrics", "the target URL")
var scrapeInterval = flag.Duration("scrape-interval", 250*time.Millisecond, "the scrape interval")
var initColumns = flag.Int("columns", 3, "the initial number of columns")
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
	win := giu.NewMasterWindow("promview: "+*targetURL, 1400, 900, 0)

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

	// prepare config
	columns := int32(*initColumns)

	// run ui code
	win.Run(func() {
		w, h := win.GetSize()
		giu.Window("Data").Pos(0, 0).Size(float32(w), float32(h)).Flags(giu.WindowFlagsNoResize|giu.WindowFlagsNoMove).Layout(
			giu.SliderInt("Columns", &columns, 1, 6),
			giu.Custom(func() {
				// prepare widgets
				var widgets []giu.Widget

				// walk series
				walk(func(s *series) {
					// prepare lists and widgets
					data := make([][]float64, 0, len(s.lists))
					lines := make([]giu.PlotWidget, 0, len(s.lists))
					for _, dim := range s.dims {
						data = append(data, s.lists[dim].slice())
						lines = append(lines, giu.PlotLine(dim, s.lists[dim].slice()))
					}

					// get min and max
					min, max := minMax(data...)

					// prepare flags
					var flags giu.PlotFlags
					if len(s.dims) == 1 && s.dims[0] == "default" {
						flags = giu.PlotFlagsNoLegend
					}

					// append widget
					widgets = append(widgets, giu.Custom(func() {
						giu.Tooltip(s.help).Build()
						giu.Plot(s.name).
							Size((w-(int(columns)*15))/int(columns), 0).
							AxisLimits(0, float64(*seriesLength), min-5, max+5, giu.ConditionAlways).
							Flags(flags).Plots(lines...).
							Build()
					}))

					// check row
					if len(widgets) == int(columns) {
						giu.Row(widgets...).Build()
						widgets = nil
					}
				})

				// check row
				if len(widgets) == int(columns) {
					giu.Row(widgets...).Build()
					widgets = nil
				}
			}),
		)
	})
}
