package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/pprof/profile"
)

var profileNodes = map[string]*node{}
var profilesMutex sync.Mutex

func loadProfile(name, url string, duration time.Duration) error {
	// get seconds
	seconds := int(duration / time.Second)
	if seconds < 1 {
		seconds = 1
	}

	// get profile
	res, err := http.Get(url + "?seconds=" + strconv.Itoa(seconds))
	if err != nil {
		return err
	}

	// ensure close
	defer res.Body.Close()

	// parse profile
	prf, err := profile.Parse(res.Body)
	if err != nil {
		return err
	}

	// prepare root
	root := &node{
		name: "root",
	}

	// convert samples
	for _, sample := range prf.Sample {
		// prepare node
		node := root

		// reverse iterate locations
		for i := len(sample.Location) - 1; i >= 0; i-- {
			// get location
			location := sample.Location[i]

			// iterate lines
			for _, line := range location.Line {
				node = node.push(line.Function.Name)
			}
		}

		// set self
		node.self += sample.Value[1]

		// increment total
		for node != nil {
			node.total += sample.Value[1]
			node = node.parent
		}
	}

	// sort nodes
	root.sort()

	// set profile
	profilesMutex.Lock()
	profileNodes[name] = root
	profilesMutex.Unlock()

	return nil
}

func getProfile(name string) *node {
	// acquire mutex
	profilesMutex.Lock()
	defer profilesMutex.Unlock()

	return profileNodes[name]
}

type walkProfileFunc func(level int, offset, length float32, name string, self, total int64)

func walkProfile(node *node, fn walkProfileFunc) {
	// get divisor
	divisor := float32(node.total)

	// walk node
	walkProfileNode(node, 0, 0, divisor, fn)
}

func walkProfileNode(nd *node, level int, offset, divisor float32, fn walkProfileFunc) float32 {
	// get length
	length := float32(nd.total) / divisor

	// emit node
	fn(level, offset, length, nd.name, nd.self, nd.total)

	// walk children
	for _, node := range nd.nodes {
		offset += walkProfileNode(node, level+1, offset, divisor, fn)
	}

	return length
}
