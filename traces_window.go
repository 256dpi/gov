package main

import (
	"image"
	"sort"
	"time"

	"github.com/AllenDang/giu"
	"github.com/samber/lo"
)

type traceWindow struct {
	name string
	open bool
}

func (w *traceWindow) draw(m *giu.MasterWindow) {
	// create window
	win := newWindow(m, w.name).IsOpen(&w.open)

	// compute keys
	traceMutex.Lock()
	keys := lo.Keys(traceStreams[w.name].events)
	traceMutex.Unlock()
	sort.Strings(keys)

	// collect rows
	var rows []*giu.TableRowWidget
	for _, task := range keys {
		task := task
		rows = append(rows, giu.TableRow(giu.Label(task), giu.Custom(func() {
			// get positions
			width, _ := giu.GetAvailableRegion()
			ratio := float64(width) / float64(traceLength.Nanoseconds())
			stop := float64(time.Now().UnixNano())
			start := stop - float64(traceLength.Nanoseconds())
			pos := giu.GetCursorPos()

			// get events
			traceMutex.Lock()
			events := traceStreams[w.name].events[task]
			traceMutex.Unlock()

			// draw events
			for _, event := range events {
				// calculate stop and start
				eventStop := (float64(event.stop.UnixNano()) - start) * ratio
				if eventStop < 0 {
					continue
				}
				eventStart := (float64(event.start.UnixNano()) - start) * ratio
				if eventStart < 0 {
					eventStart = 0
				}

				// draw progress bar
				giu.SetCursorPos(image.Pt(pos.X+int(eventStart), pos.Y))
				giu.ProgressBar(1).Size(float32(eventStop-eventStart), 20).Build()
			}
		})))
	}

	// draw
	win.Layout(
		giu.Table().Columns(
			giu.TableColumn("Task").Flags(giu.TableColumnFlagsWidthFixed).InnerWidthOrWeight(200),
			giu.TableColumn("Calls").Flags(giu.TableColumnFlagsWidthStretch).InnerWidthOrWeight(1),
		).Rows(rows...),
	)
}
