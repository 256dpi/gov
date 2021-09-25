package main

import "math"

type list struct {
	length    int
	data      []float64
	lastItem  int
	lastValue float64
}

func newList(length int) *list {
	return &list{
		length: length,
		data:   make([]float64, length*2),
	}
}

func (l *list) add(value float64, diff bool) {
	// increment position
	l.lastItem++
	if l.lastItem >= l.length {
		l.lastItem = 0
	}

	// get difference
	n := value
	if diff {
		n -= l.lastValue
	}

	// write values
	l.data[l.lastItem] = n
	l.data[l.length+l.lastItem] = n
	l.lastValue = value
}

func (l *list) slice() []float64 {
	return l.data[l.lastItem : l.lastItem+l.length]
}

func extent(lists ...[]float64) int {
	// find extent
	n := 0
	for _, list := range lists {
		if len(list) > n {
			n = len(list)
		}
	}

	return n
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
