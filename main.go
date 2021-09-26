package main

import (
	"flag"
	"image"
	"image/color"
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

func profiles(mw *giu.MasterWindow) func() {
	return func() {
		// get size
		mw, mh := mw.GetSize()

		// create window
		win := giu.Window("Profile")
		win.Pos(100, 100)
		win.Size(float32(mw)*0.7, float32(mh)*0.7)

		// prepare colors
		text := color.RGBA{R: 242, G: 245, B: 250, A: 255}
		fill := color.RGBA{R: 51, G: 64, B: 74, A: 255}
		line := color.RGBA{R: 26, G: 33, B: 39, A: 255}

		// get positions and size
		_x, _y := win.CurrentPosition()
		x, y := int(_x), int(_y)
		w, _ := win.CurrentSize()

		// adjust position and size
		x += 10
		y += 30
		w -= 20

		// get mouse
		mouse := giu.GetMousePos()

		// draw
		win.Layout(
			giu.Custom(func() {
				// get canvas
				canvas := giu.GetCanvas()

				// walk profile
				walkProfile(func(level int, offset, length float32, name string, value int64) {
					// get dimensions
					p1x := x + int(offset*w)
					p1y := y + level*50
					p2x := x + int((offset+length)*w)
					p2y := y + level*50 + 50

					// determine hover
					hover := mouse.In(image.Rect(p1x, p1y, p2x, p2y))

					// draw item
					canvas.AddRectFilled(image.Pt(p1x, p1y), image.Pt(p2x, p2y), fill, 0, 0)
					canvas.AddRect(image.Pt(p1x, p1y), image.Pt(p2x, p2y), line, 0, 0, 1)
					canvas.AddText(image.Pt(p1x+5, p1y+5), text, name)
					canvas.AddText(image.Pt(p1x+5, p1y+25), text, time.Duration(value).String())

					// handle hover
					if hover {
						giu.Tooltip(name).Build()
					}
				})
			}),
		)
	}
}
