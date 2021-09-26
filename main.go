package main

import (
	"flag"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/AllenDang/giu"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var seriesLength = flag.Int("series-length", 100, "the series length")
var targetURL = flag.String("target-url", "http://0.0.0.0:6060/", "the target URL")
var metricsPath = flag.String("metrics-path", "metrics", "the metrics path")
var profilePath = flag.String("profile-path", "profile", "the profile path")
var scrapeInterval = flag.Duration("scrape-interval", 250*time.Millisecond, "the scrape interval")
var initColumns = flag.Int("columns", 3, "the initial number of columns")
var metricsAddr = flag.String("metrics-addr", ":6060", "the UI metrics addr")

func main() {
	// parse flags
	flag.Parse()

	// run prometheus and pprof profile endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/profile", pprof.Profile)
	go func() {
		panic(http.ListenAndServe(*metricsAddr, nil))
	}()

	// create window
	mw := giu.NewMasterWindow("promview: "+*targetURL, 1400, 900, 0)

	// run scraper
	go func() {
		for {
			// scrape metric families
			families, err := scrape(*targetURL + *metricsPath)
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

	// run profiler
	go func() {
		for {
			_, err := getProfile(*targetURL + *profilePath)
			if err != nil {
				panic(err)
			}
		}
	}()

	// get drawers
	drawMetrics := metrics(mw)

	// run ui code
	mw.Run(func() {
		drawMetrics()
	})
}

func metrics(mw *giu.MasterWindow) func() {
	// prepare config
	columns := int32(*initColumns)

	return func() {
		// get size
		mw, mh := mw.GetSize()

		// create window
		win := giu.Window("Metrics")
		win.Pos(100, 100)
		win.Size(float32(mw)*0.7, float32(mh)*0.7)

		// get current size
		w, _ := win.CurrentSize()

		win.Layout(
			// add columns slider
			giu.SliderInt("Columns", &columns, 1, 6),

			// add plots
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
							Size((int(w)-(int(columns)*15))/int(columns), 0).
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
	}
}
