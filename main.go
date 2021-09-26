package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/AllenDang/giu"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var seriesLength = flag.Int("series-length", 100, "the series length")
var targetURL = flag.String("target-url", "http://0.0.0.0:6060/", "the target URL")
var metricsPath = flag.String("metrics-path", "metrics", "the metrics path")
var profilePath = flag.String("profile-path", "debug/pprof/profile", "the profile path")
var scrapeInterval = flag.Duration("scrape-interval", 250*time.Millisecond, "the scrape interval")
var initColumns = flag.Int("columns", 3, "the initial number of columns")
var selfAddr = flag.String("self-addr", ":7070", "the UI metrics addr")

func main() {
	// parse flags
	flag.Parse()

	// run prometheus and pprof profile endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		panic(http.ListenAndServe(*selfAddr, nil))
	}()

	// create window
	mw := giu.NewMasterWindow("gov: "+*targetURL, 1400, 900, 0)

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
			prf, err := getProfile(*targetURL + *profilePath)
			if err != nil {
				panic(err)
			}

			convertProfile(prf)
		}
	}()

	// get drawers
	drawMetrics := metrics(mw)
	drawProfiles := profiles(mw)

	// run ui code
	mw.Run(func() {
		// background
		gl.ClearColor(40.0/255.0, 45.0/255.0, 50.0/255.0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// draw widgets
		drawMetrics()
		drawProfiles()
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
						giu.Plot(s.name).
							Size((int(w)-(int(columns)*15))/int(columns), 0).
							AxisLimits(0, float64(*seriesLength), min-5, max+5, giu.ConditionAlways).
							Flags(flags).Plots(lines...).
							Build()
						giu.Tooltip(s.help).Build()
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

func profiles(mw *giu.MasterWindow) func() {
	return func() {
		// get size
		mw, mh := mw.GetSize()

		// create window
		win := giu.Window("Profile")
		win.Pos(100, 100)
		win.Size(float32(mw)*0.7, float32(mh)*0.7)

		// get positions and size
		w, _ := win.CurrentSize()
		x := 10
		y := 30
		w -= 30

		// draw
		win.Layout(
			giu.Custom(func() {
				// override style
				giu.PushStyleColor(giu.StyleColorProgressBarActive, color.RGBA{R: 58, G: 82, B: 99, A: 255})
				defer giu.PopStyleColor()

				// walk profile
				walkProfile(func(level int, offset, length float32, name string, self, total int64) {
					// set cursor
					giu.SetCursorPos(image.Pt(x+int(offset*w), y+level*50))

					// get text
					text := fmt.Sprintf("%s (%s/%s)", name, time.Duration(self).String(), time.Duration(total).String())

					// build tooltip and button
					// giu.Button(text).Size(length*w, 50).Build()
					giu.ProgressBar(float32(self)/float32(total)).Size(length*w, 50).Overlay(text).Build()
					giu.Tooltip(text).Build()
				})
			}),
		)
	}
}
