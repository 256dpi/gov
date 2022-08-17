package main

import (
	"github.com/AllenDang/giu"
	"github.com/dustin/go-humanize"
)

type metricWindow struct {
	node *metricsNode
	cols int32
	open bool
}

func (w *metricWindow) draw(m *giu.MasterWindow) {
	// create window
	win := newWindow(m, w.node.name).Flags(giu.WindowFlagsMenuBar).IsOpen(&w.open)

	// get size
	width, _ := win.CurrentSize()

	win.Layout(
		// add menu bar
		giu.MenuBar().Layout(
			giu.Menu("View").Layout(
				giu.SliderInt(&w.cols, 1, 4).Label("Columns"),
			),
		),

		// add plots
		giu.Custom(func() {
			// prepare widgets
			var widgets []giu.Widget

			// walk metrics
			walkMetrics(w.node, func(s *metricSeries) {
				// prepare lists and widgets
				data := make([][]float64, 0, len(s.lists))
				lines := make([]giu.PlotWidget, 0, len(s.lists))
				for _, dim := range s.dims {
					data = append(data, s.lists[dim].slice())
					lines = append(lines, giu.PlotLine(dim, s.lists[dim].slice()))
				}

				// get min and max
				min, max := minMax(data...)
				r := (max - min) * 0.1
				min -= r
				max += r

				// prepare flags
				var flags giu.PlotFlags
				if len(s.dims) == 1 && s.dims[0] == "default" {
					flags = giu.PlotFlagsNoLegend
				}

				// generate tick values
				r = max - min
				ticks := []giu.PlotTicker{
					{Position: min},
					{Position: min + r/3},
					{Position: min + r/3*2},
					{Position: max},
				}

				// set labels
				for i := range ticks {
					ticks[i].Label = humanize.SIWithDigits(ticks[i].Position, 2, "")
				}

				// append widget
				widgets = append(widgets, giu.Custom(func() {
					giu.Plot(s.name).
						Size((int(width)-20-(int(w.cols)*8))/int(w.cols), 0).
						AxisLimits(0, float64(*seriesLength), min, max, giu.ConditionAlways).
						XAxeFlags(giu.PlotAxisFlagsNoTickLabels).
						YTicks(ticks, false, 0).
						Flags(flags).Plots(lines...).
						Build()
					giu.Tooltip(s.help).Build()
				}))

				// check row
				if len(widgets) == int(w.cols) {
					giu.Row(widgets...).Build()
					widgets = nil
				}
			})

			// check row
			if len(widgets) > 0 {
				giu.Row(widgets...).Build()
			}
		}),
	)
}
