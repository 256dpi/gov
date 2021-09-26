package main

import "sort"

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
