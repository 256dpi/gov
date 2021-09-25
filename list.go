package main

import "math"

type list struct {
	length int
	data   []float64
	pos    int
	last   float64
}

func newList(length int) *list {
	return &list{
		length: length,
		data:   make([]float64, length*2),
	}
}

func (l *list) add(value float64, diff bool) {
	// get difference
	n := value
	if diff {
		n -= l.last
	}

	// write values
	l.data[l.pos] = n
	l.data[l.length+l.pos] = n
	l.last = value

	// increment position
	l.pos++
	if l.pos >= l.length {
		l.pos = 0
	}
}

func (l *list) slice() []float64 {
	return l.data[l.pos : l.length+l.pos]
}

func minMax(lists ...[]float64) (float64, float64) {
	// find minimum and maximum
	min, max := lists[0][0], lists[0][0]
	for _, list := range lists {
		for _, value := range list {
			min = math.Min(min, value)
			max = math.Max(max, value)
		}
	}

	return min, max
}
