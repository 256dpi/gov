package main

import "time"

type traceEvent struct {
	start time.Time
	stop  time.Time
}

type traceStream struct {
	events map[string][]traceEvent
}
