package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/AllenDang/giu"
)

type profileWindow struct {
	name    string
	title   string
	open    bool
	stream  bool
	profile *node
}

func (w *profileWindow) update() {
	// update profile
	if w.stream {
		w.profile = getProfile(w.name)
	}
}

func (w *profileWindow) draw(mw *giu.MasterWindow) {
	// create window
	win := newWindow(mw, w.title).Flags(giu.WindowFlagsMenuBar).IsOpen(&w.open)

	// get position and size
	width, _ := win.CurrentSize()
	posX := 10
	posY := 50
	width -= 30

	// draw
	win.Layout(
		giu.MenuBar().Layout(
			giu.Condition(w.stream, giu.Layout{
				giu.MenuItem("Pause Stream").OnClick(func() {
					w.stream = false
				}),
			}, giu.Layout{
				giu.MenuItem("Start Stream").OnClick(func() {
					w.stream = true
				}),
			}),
		),

		giu.Custom(func() {
			// check profile
			if w.profile == nil {
				return
			}

			// override style
			giu.PushStyleColor(giu.StyleColorProgressBarActive, color.RGBA{R: 58, G: 82, B: 99, A: 255})
			defer giu.PopStyleColor()

			// walk profile
			walkProfile(w.profile, func(level int, offset, length float32, name string, self, total int64) {
				// set cursor
				giu.SetCursorPos(image.Pt(posX+int(offset*width), posY+level*30))

				// get text
				text := fmt.Sprintf("%s (%s/%s)", name, time.Duration(self).String(), time.Duration(total).String())

				// build tooltip and button
				giu.ProgressBar(float32(self)/float32(total)).Size(length*width, 30).Overlay(text).Build()
				giu.Tooltip(text).Build()
			})
		}),
	)
}
