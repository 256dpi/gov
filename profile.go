package main

import (
	"net/http"
	"sort"
	"sync"

	"github.com/google/pprof/profile"
)

var lastNode *node
var lastNodeMutex sync.Mutex

type node struct {
	name   string
	self   int64
	total  int64
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

func (n *node) sort() {
	// sort nodes
	sort.Slice(n.nodes, func(i, j int) bool {
		return n.nodes[i].name < n.nodes[j].name
	})

	// descend
	for _, node := range n.nodes {
		node.sort()
	}
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
		name: "#root",
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

	// set
	lastNodeMutex.Lock()
	lastNode = root
	lastNodeMutex.Unlock()

	return root
}

type walkFN func(level int, offset, length float32, name string, self, total int64)

func walkProfile(fn walkFN) {
	// acquire mutex
	lastNodeMutex.Lock()
	defer lastNodeMutex.Unlock()

	// walk last root node
	if lastNode != nil {
		// get divisor
		divisor := float32(lastNode.total)

		walkNode(lastNode, 0, 0, divisor, fn)
	}
}

func walkNode(nd *node, level int, offset, divisor float32, fn walkFN) float32 {
	// get length
	length := float32(nd.total) / divisor

	// emit node
	fn(level, offset, length, nd.name, nd.self, nd.total)

	// walk children
	for _, node := range nd.nodes {
		offset += walkNode(node, level+1, offset, divisor, fn)
	}

	return length
}
