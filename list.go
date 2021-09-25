package main

import "math"

const storage = 100

type list struct {
	data      [storage * 2]float64
	lastValue float64
	lastPos   int
	length    int
}

func newList() *list {
	return &list{}
}

func (l *list) add(value float64, diff bool) {
	// increment position
	l.lastPos++
	if l.lastPos >= storage {
		l.lastPos = 0
	}

	// get difference
	n := value
	if diff {
		n -= l.lastValue
	}

	// write values
	l.data[l.lastPos] = n
	l.data[storage+l.lastPos] = n
	l.lastValue = value

	// increment length
	if l.length < storage {
		l.length++
	}
}

func (l *list) slice() []float64 {
	return l.data[l.lastPos : l.lastPos+l.length]
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
