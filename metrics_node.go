package main

import "sort"

type metricsNode struct {
	name     string
	series   *metricSeries
	parent   *metricsNode
	children []*metricsNode
}

func (n *metricsNode) ensure(path []string) *metricsNode {
	// prepare node
	node := n

	// push nodes
	for len(path) > 0 {
		node = node.push(path[0])
		path = path[1:]
	}

	return node
}

func (n *metricsNode) push(name string) *metricsNode {
	// check children
	for _, child := range n.children {
		if child.name == name {
			return child
		}
	}

	// create child
	child := &metricsNode{
		name:   name,
		parent: n,
	}

	// add child
	n.children = append(n.children, child)

	// sort children
	sort.Slice(n.children, func(i, j int) bool {
		return n.children[i].name < n.children[j].name
	})

	return child
}

func (n *metricsNode) walk(fn func(*metricsNode)) {
	// emit self
	fn(n)

	// walk children
	for _, child := range n.children {
		child.walk(fn)
	}
}
