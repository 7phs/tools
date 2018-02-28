package common

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	tryRandomCount = 10
	infiniteTimeout = 1000000 * time.Hour
)

type randomRange struct {
	checked map[int]bool
	min     int
	delta   int
}

func RandomRange(min, max int, values ... int) *randomRange {
	min, max = minMax(min, max)

	return (&randomRange{
		checked: make(map[int]bool),
		min:     min,
		delta:   max - min,
	}).init(values...)
}

func (o *randomRange) init(values ... int) *randomRange {
	for value := range values {
		o.checked[value] = true
	}

	return o
}

func (o *randomRange) Int() (value int) {
	exists := true

	for i := 0; exists && i < tryRandomCount; i++ {
		value = o.min + rand.Intn(o.delta)

		if _, exists = o.checked[value]; !exists {
			o.checked[value] = true
		}
	}

	return
}

func minMax(min, max int) (int, int) {
	if min < max {
		return min, max
	} else {
		return max, min
	}
}

func isOutOfRange(value, min, max int) bool {
	return value<min || value>max
}

func rangeMinMax(defMin, defMax int, minMaximum ... int) (int, int) {
	min, max := func(values ... int) (int, int) {
		return values[0], values[1]
	}(takeArgs(2, minMaximum...)...)

	if isOutOfRange(min, defMin, defMax) {
		min = defMin
	}

	if isOutOfRange(max, defMin, defMax) {
		max = defMax
	}

	return minMax(min, max)
}

func takeArgs(ln int, minMaximum ... int) []int {
	if delta := ln-len(minMaximum); delta>0 {
		minMaximum = append(minMaximum, make([]int, delta)...)
	}

	return minMaximum[:ln]
}

func isPortAvailable(port int) bool {
	if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		return false
	} else {
		ln.Close()
	}

	return true
}

func CheckLocalPort(port int, minMaximum ... int) int {
	min, max := rangeMinMax(1000, 64000, minMaximum...)

	random := RandomRange(min, max, port)

	for i := 0; !isPortAvailable(port) && i < tryRandomCount; i++ {
		port = random.Int()
	}

	return port
}

func DeadLineTimer(deadline time.Time, ok bool) *time.Timer {
	var duration time.Duration

	if !ok {
		duration = infiniteTimeout
	} else if duration = deadline.Sub(time.Now()); duration<0 {
		duration = 0
	}

	return time.NewTimer(duration)
}