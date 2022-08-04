package main

import "math"

type list struct {
	length    int
	data      []float64
	pos       int
	lastValue float64
	lastSum   float64
	lastCount float64
}

func newList(length int) *list {
	return &list{
		length: length,
		data:   make([]float64, length*2),
	}
}

func (l *list) addDiff(value float64) {
	l.add(value - l.lastValue)
	l.lastValue = value
}

func (l *list) addMean(sum, count float64) {
	l.add((sum - l.lastSum) / (count - l.lastCount))
	l.lastSum = sum
	l.lastCount = count
}

func (l *list) add(value float64) {
	// handle not a number
	if math.IsNaN(value) {
		value = 0
	}

	// write values
	l.data[l.pos] = value
	l.data[l.length+l.pos] = value

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
