package main

import (
	"fmt"
	"strconv"

	"github.com/AllenDang/giu"
)

func newWindow(m *giu.MasterWindow, title string) *giu.WindowWidget {
	// get size
	mw, mh := m.GetSize()

	// create window
	win := giu.Window(title)
	win.Pos(100, 100)
	win.Size(float32(mw-200), float32(mh-200))

	return win
}

func fmtBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func f2s(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
