package main

import (
	"bufio"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
)

const traceLength = 10 * time.Second

var traceStreams = map[string]*traceStream{}
var traceMutex sync.Mutex

func loadTraces(url string, refresh func()) error {
	// open stream
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// scan stream
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		// split line
		seg := strings.Split(scanner.Text(), ";")
		if len(seg) != 4 {
			continue
		}

		// get fields
		name := seg[0]
		task := seg[1]
		start, _ := time.Parse(time.RFC3339Nano, seg[2])
		stop, _ := time.Parse(time.RFC3339Nano, seg[3])

		// add event
		traceMutex.Lock()
		stream := traceStreams[name]
		if stream == nil {
			stream = &traceStream{
				events: map[string][]traceEvent{},
			}
			traceStreams[name] = stream
		}
		stream.events[task] = append(stream.events[task], traceEvent{
			start: start,
			stop:  stop,
		})
		max := time.Now().Add(-traceLength)
		stream.events[task] = lo.Filter(stream.events[task], func(event traceEvent, i int) bool {
			return event.stop.After(max)
		})
		traceMutex.Unlock()

		// refresh
		refresh()
	}

	return scanner.Err()
}
