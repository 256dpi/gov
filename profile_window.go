package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/AllenDang/giu"
)

type profileWindow struct {
	name  string
	title string
	open  bool
}

func (w *profileWindow) draw(mw *giu.MasterWindow) {
	// create window
	win := newWindow(mw, w.title).IsOpen(&w.open)

	// get position and size
	width, _ := win.CurrentSize()
	posX := 10
	posY := 30
	width -= 30

	// draw
	win.Layout(
		giu.Custom(func() {
			// override style
			giu.PushStyleColor(giu.StyleColorProgressBarActive, color.RGBA{R: 58, G: 82, B: 99, A: 255})
			defer giu.PopStyleColor()

			// walk profile
			walkProfile(w.name, func(level int, offset, length float32, name string, self, total int64) {
				// set cursor
				giu.SetCursorPos(image.Pt(posX+int(offset*width), posY+level*30))

				// get text
				text := fmt.Sprintf("%s (%s/%s)", name, time.Duration(self).String(), time.Duration(total).String())

				// build tooltip and button
				// giu.Button(text).Size(length*w, 50).Build()
				giu.ProgressBar(float32(self)/float32(total)).Size(length*width, 30).Overlay(text).Build()
				giu.Tooltip(text).Build()
			})
		}),
	)
}
