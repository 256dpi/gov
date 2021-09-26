package main

import "github.com/AllenDang/giu"

func newWindow(m *giu.MasterWindow, title string) *giu.WindowWidget {
	// get size
	mw, mh := m.GetSize()

	// create window
	win := giu.Window(title)
	win.Pos(100, 100)
	win.Size(float32(mw-200), float32(mh-200))

	return win
}
