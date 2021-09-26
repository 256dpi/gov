package main

import (
	"flag"
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
var cpuProfilePath = flag.String("cpu-profile-path", "debug/pprof/profile", "the CPU profile path")
var heapProfilePath = flag.String("heap-profile-path", "debug/pprof/heap", "the heap profile path")
var scrapeInterval = flag.Duration("scrape-interval", 250*time.Millisecond, "the scrape interval")
var initColumns = flag.Int("columns", 3, "the initial number of columns")
var selfAddr = flag.String("self-addr", ":7070", "the UI metrics addr")

var metricWindows = map[string]*metricWindow{}
var profileWindows = map[string]*profileWindow{}

func main() {
	// parse flags
	flag.Parse()

	// run prometheus and pprof profile endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		panic(http.ListenAndServe(*selfAddr, nil))
	}()

	// create master window
	master := giu.NewMasterWindow(*targetURL, 1400, 900, 0)

	// run metrics loader
	go metricsLoader(*targetURL+*metricsPath, *scrapeInterval)

	// run profiler loaders
	go profileLoader("cpu", *targetURL+*cpuProfilePath)
	go profileLoader("heap", *targetURL+*heapProfilePath)
	go profileLoader("block", *targetURL+*heapProfilePath)
	go profileLoader("mutex", *targetURL+*heapProfilePath)

	// run ui code
	master.Run(func() {
		// main menu
		withMetricsTree(func(tree *metricsNode) {
			giu.MainMenuBar().Layout(
				giu.Menu("Metrics").Layout(
					buildMetricsMenu(tree),
				),
				giu.Menu("Profiles").Layout(
					buildProfileMenuItem("cpu", "CPU"),
					buildProfileMenuItem("heap", "Heap"),
					buildProfileMenuItem("block", "Block"),
					buildProfileMenuItem("mutex", "Mutex"),
				),
			).Build()
		})

		// background
		gl.ClearColor(40.0/255.0, 45.0/255.0, 50.0/255.0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// draw metric windows
		for key, win := range metricWindows {
			if !win.open {
				delete(metricWindows, key)
			} else {
				win.draw(master)
			}
		}

		// draw profile windows
		for key, win := range profileWindows {
			if !win.open {
				delete(profileWindows, key)
			} else {
				win.draw(master)
			}
		}
	})
}

func buildMetricsMenu(node *metricsNode) giu.Widget {
	// prepare click handler
	click := func() {
		if metricWindows[node.name] == nil {
			metricWindows[node.name] = &metricWindow{
				node: node,
				cols: int32(*initColumns),
				open: true,
			}
		}
	}

	// check children
	if len(node.children) == 0 {
		return giu.MenuItem(node.name).OnClick(click)
	}

	// prepare widgets
	widgets := make([]giu.Widget, 0, 1+len(node.children))

	// add show
	widgets = append(widgets, giu.MenuItem("Show").OnClick(click))

	// add children
	for _, child := range node.children {
		widgets = append(widgets, buildMetricsMenu(child))
	}

	return giu.Menu(node.name).Layout(widgets...)
}

func buildProfileMenuItem(name, title string) *giu.MenuItemWidget {
	return giu.MenuItem(title).OnClick(func() {
		if profileWindows[name] == nil {
			profileWindows[name] = &profileWindow{
				name:  name,
				title: title,
				open:  true,
			}
		}
	})
}

func metricsLoader(url string, interval time.Duration) {
	for {
		// scrape metric
		err := scrapeMetrics(url)
		if err != nil {
			println("metrics: " + err.Error())
		}

		// update
		giu.Update()

		// await next interval
		time.Sleep(interval)
	}
}

func profileLoader(name, url string) {
	for {
		// load profile
		err := loadProfile(name, url)
		if err != nil {
			println("profile: " + err.Error())
		}

		// update
		giu.Update()
	}
}
