package main

import (
	"net/http"
	"sync"

	"github.com/google/pprof/profile"
)

var lastNode *node
var lastNodeMutex sync.Mutex

type node struct {
	name   string
	value  int64
	parent *node
	nodes  []*node
}

func (n *node) push(name string) *node {
	// check nodes
	for _, node := range n.nodes {
		if node.name == name {
			return node
		}
	}

	// create node
	node := &node{
		name:   name,
		parent: n,
	}

	// add node
	n.nodes = append(n.nodes, node)

	return node
}

func getProfile(url string) (*profile.Profile, error) {
	// get profile
	res, err := http.Get(url + "?seconds=1")
	if err != nil {
		return nil, err
	}

	// ensure close
	defer res.Body.Close()

	// parse profile
	prf, err := profile.Parse(res.Body)
	if err != nil {
		return nil, err
	}

	return prf, nil
}

func convertProfile(prf *profile.Profile) *node {
	// prepare root
	root := &node{
		name:  "#root",
		value: 1,
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

		// increment value
		for node != nil {
			node.value += sample.Value[1]
			node = node.parent
		}
	}

	// set
	lastNodeMutex.Lock()
	lastNode = root
	lastNodeMutex.Unlock()

	return root
}

type walkFN func(level int, offset, length float32, name string, value int64)

func walkProfile(fn walkFN) {
	// acquire mutex
	lastNodeMutex.Lock()
	defer lastNodeMutex.Unlock()

	// walk last root node
	if lastNode != nil {
		// get divisor
		divisor := float32(lastNode.value)

		walkNode(lastNode, 0, 0, divisor, fn)
	}
}

func walkNode(nd *node, level int, offset, divisor float32, fn walkFN) float32 {
	// get length
	length := float32(nd.value) / divisor

	// emit node
	fn(level, offset, length, nd.name, nd.value)

	// walk children
	for _, node := range nd.nodes {
		offset += walkNode(node, level+1, offset, divisor, fn)
	}

	return length
}
