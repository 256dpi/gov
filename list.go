package main

import "math"

const storage = 100

type list struct {
	data   [storage * 2]float64
	last   int
	length int
}

func newList() *list {
	return &list{}
}

func (l *list) add(n float64) {
	// increment position
	l.last++
	if l.last >= storage {
		l.last = 0
	}

	// write values
	l.data[l.last] = n
	l.data[storage+l.last] = n

	// increment length
	if l.length < storage {
		l.length++
	}
}

func (l *list) slice() []float64 {
	return l.data[l.last : l.last+l.length]
}

func minMax(list []float64) (float64, float64) {
	// find minimum and maximum
	min, max := list[0], list[0]
	for _, value := range list {
		min = math.Min(min, value)
		max = math.Max(max, value)
	}

	return min, max
}
