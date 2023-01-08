package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
)

var seriesLength = flag.Int("series-length", 100, "the series length")
var metricsPath = flag.String("metrics-path", "/metrics", "the metrics path")
var tracePath = flag.String("traces-path", "/trace", "the trace path")
var cpuProfilePath = flag.String("cpu-profile-path", "/debug/pprof/profile", "the CPU profile path")
var allocsProfilePath = flag.String("allocs-profile-path", "/debug/pprof/allocs", "the allocs profile path")
var heapProfilePath = flag.String("heap-profile-path", "/debug/pprof/heap", "the heap profile path")
var blockProfilePath = flag.String("block-profile-path", "/debug/pprof/block", "the block profile path")
var mutexProfilePath = flag.String("mutex-profile-path", "/debug/pprof/mutex", "the mutex profile path")
var scrapeInterval = flag.Duration("scrape-interval", 250*time.Millisecond, "the default scrape interval")
var profileInterval = flag.Duration("profile-interval", 2*time.Second, "the default profile interval")
var initColumns = flag.Int("columns", 3, "the default number of columns")
var selfAddr = flag.String("self-addr", ":7070", "the address for govs own metrics")
var metricsSplitDepth = flag.Int("metrics-split-depth", 3, "the metrics split depth")

var metricWindows = map[string]*metricWindow{}
var traceWindows = map[string]*traceWindow{}
var profileWindows = map[string]*profileWindow{}

var autoUpdate = false

func main() {
	// parse flags
	flag.Parse()

	// run prometheus and pprof profile endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		panic(http.ListenAndServe(*selfAddr, nil))
	}()

	// get target
	target := strings.TrimRight(flag.Arg(0), "/")
	if target == "" {
		target = "http://0.0.0.0:6060"
	}

	// create master window
	master := giu.NewMasterWindow(target, 1400, 900, 0)

	// allow long draw lists
	imgui.CurrentIO().SetBackendFlags(imgui.BackendFlagsRendererHasVtxOffset)

	// run metrics and trace loader
	go metricsLoader(target + *metricsPath)
	go traceLoader(target + *tracePath)

	// run profiler loaders
	go profileLoader("cpu", "cpu", target+*cpuProfilePath)
	go profileLoader("allocs", "alloc_space", target+*allocsProfilePath)
	go profileLoader("heap", "inuse_space", target+*heapProfilePath)
	go profileLoader("block", "delay", target+*blockProfilePath)
	go profileLoader("mutex", "delay", target+*mutexProfilePath)

	// prepare scrape intervals
	scrapeIntervals := []time.Duration{
		100 * time.Millisecond,
		250 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	// prepare profile intervals
	profileIntervals := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}

	// run periodic updater (50 Hz)
	go func() {
		for range time.Tick(20 * time.Millisecond) {
			if autoUpdate {
				giu.Update()
			}
		}
	}()

	// run ui code
	master.Run(func() {
		// update profile windows
		for _, win := range profileWindows {
			win.update()
		}

		/* draw */

		// main menu
		withMetricsTree(func(tree *metricsNode) {
			giu.MainMenuBar().Layout(
				giu.Menu("Metrics").Layout(
					buildMetricsMenuItems(tree)...,
				),
				giu.Menu("Traces").Layout(
					buildTracesMenuItems()...,
				),
				giu.Menu("Profiles").Layout(
					buildProfileMenuItem("cpu", "CPU"),
					buildProfileMenuItem("allocs", "Allocs"),
					buildProfileMenuItem("heap", "Heap"),
					buildProfileMenuItem("block", "Block"),
					buildProfileMenuItem("mutex", "Mutex"),
				),
				giu.Menu("Settings").Layout(
					giu.Menu("Scrape Interval").Layout(
						lo.Map(scrapeIntervals, func(interval time.Duration, _ int) giu.Widget {
							return giu.MenuItem(interval.String()).Selected(*scrapeInterval == interval).OnClick(func() {
								*scrapeInterval = interval
							})
						})...,
					),
					giu.Menu("Profile Interval").Layout(
						lo.Map(profileIntervals, func(interval time.Duration, _ int) giu.Widget {
							return giu.MenuItem(interval.String()).Selected(*profileInterval == interval).OnClick(func() {
								*profileInterval = interval
							})
						})...,
					),
				),
			).Build()
		})

		// background
		gl.ClearColor(40.0/255.0, 45.0/255.0, 50.0/255.0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// reset auto update
		autoUpdate = false

		// draw metric windows
		for key, win := range metricWindows {
			if !win.open {
				delete(metricWindows, key)
			} else {
				win.draw(master)
			}
		}

		// draw metric windows
		for key, win := range traceWindows {
			if !win.open {
				delete(traceWindows, key)
			} else {
				autoUpdate = true
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

func buildMetricsMenuItems(node *metricsNode) []giu.Widget {
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
		return []giu.Widget{giu.MenuItem(node.name).OnClick(click)}
	}

	// prepare widgets
	widgets := make([]giu.Widget, 0, 1+len(node.children))

	// add show
	widgets = append(widgets, giu.MenuItem("Show All").OnClick(click))

	// add children
	for _, child := range node.children {
		// get child items
		items := buildMetricsMenuItems(child)

		// add menu or item
		if len(items) > 1 {
			widgets = append(widgets, giu.Menu(child.name).Layout(items...))
		} else {
			widgets = append(widgets, items...)
		}
	}

	return widgets
}

func buildTracesMenuItems() []giu.Widget {
	// collect widgets
	var widgets []giu.Widget
	traceMutex.Lock()
	for name := range traceStreams {
		name := name
		widgets = append(widgets, giu.MenuItem(name).OnClick(func() {
			if traceWindows[name] == nil {
				traceWindows[name] = &traceWindow{
					name: name,
					open: true,
				}
			}
		}))
	}
	traceMutex.Unlock()

	return widgets
}

func buildProfileMenuItem(name, title string) *giu.MenuItemWidget {
	return giu.MenuItem(title).OnClick(func() {
		if profileWindows[name] == nil {
			profileWindows[name] = &profileWindow{
				name:   name,
				title:  title,
				open:   true,
				stream: true,
			}
		}
	})
}

func metricsLoader(url string) {
	for {
		// scrape metric
		err := scrapeMetrics(url, *metricsSplitDepth)
		if err != nil {
			println("metrics: " + err.Error())
		}

		// update
		giu.Update()

		// await next interval
		time.Sleep(*scrapeInterval)
	}
}

func traceLoader(url string) {
	for {
		// load traces
		err := loadTraces(url, func() {
			giu.Update()
		})
		if err != nil {
			println("trace: " + err.Error())
		}

		// debounce reconnect
		time.Sleep(time.Second)
	}
}

func profileLoader(name, sample, url string) {
	for {
		// check window
		if profileWindows[name] == nil {
			time.Sleep(*profileInterval)
			continue
		}

		// load profile
		err := loadProfile(name, sample, url, *profileInterval)
		if err != nil {
			println("profile: " + err.Error())
		}

		// update
		giu.Update()
	}
}
