package main

import "github.com/AllenDang/giu"

type metricWindow struct {
	node *metricsNode
	cols int32
	open bool
}

func (w *metricWindow) draw(m *giu.MasterWindow) {
	// create window
	win := newWindow(m, w.node.name).IsOpen(&w.open)

	// get size
	width, _ := win.CurrentSize()

	win.Layout(
		// add columns slider
		giu.SliderInt("Columns", &w.cols, 1, 6),

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

				// prepare flags
				var flags giu.PlotFlags
				if len(s.dims) == 1 && s.dims[0] == "default" {
					flags = giu.PlotFlagsNoLegend
				}

				// append widget
				widgets = append(widgets, giu.Custom(func() {
					giu.Plot(s.name).
						Size((int(width)-(int(w.cols)*15))/int(w.cols), 0).
						AxisLimits(0, float64(*seriesLength), min-5, max+5, giu.ConditionAlways).
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
