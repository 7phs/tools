package common

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestMinMax(t *testing.T) {
	type testPair struct {
		min int
		max int
	}

	TestPair := func(min, max int) testPair {
		return testPair{
			min: min,
			max: max,
		}
	}

	suites := []struct {
		exist    testPair
		expected testPair
	}{
		{testPair{min: 100, max: 200}, testPair{min: 100, max: 200}},
		{testPair{min: 200, max: 100}, testPair{min: 100, max: 200}},
		{testPair{min: 200, max: 200}, testPair{min: 200, max: 200}},
	}

	for _, test := range suites {
		if exist := TestPair(minMax(test.exist.min, test.exist.max)); exist != test.expected {
			t.Error("failed to sort min and max. Got ", exist, ", but expected is ", test.expected)
		}
	}
}

func TestIsOutOfRange(t *testing.T) {
	if value, min, max := 100, 10, 1000; isOutOfRange(value, min, max) {
		t.Error("failed to check range againsts ", value, " and [", min, ",", max, "]")
	}

	if value, min, max := 5, 10, 1000; !isOutOfRange(value, min, max) {
		t.Error("failed to check range againsts ", value, " and [", min, ",", max, "]")
	}

	if value, min, max := 5000, 10, 1000; !isOutOfRange(value, min, max) {
		t.Error("failed to check range againsts ", value, " and [", min, ",", max, "]")
	}
}

func TestRangeMinMax(t *testing.T) {
	const (
		defMin = 100
		defMax = 1000
	)

	type resultPair struct {
		min int
		max int
	}

	ResultPair := func(min, max int) resultPair {
		return resultPair{
			min: min,
			max: max,
		}
	}

	type testPair struct {
		defMin     int
		defMax     int
		minMaximum []int
	}

	TestPair := func(minMaximum ... int) testPair {
		return testPair{
			defMin:     defMin,
			defMax:     defMax,
			minMaximum: minMaximum,
		}
	}

	suites := []struct {
		exist    testPair
		expected resultPair
	}{
		{TestPair(), resultPair{min: defMin, max: defMax}},
		{TestPair(200), resultPair{min: 200, max: defMax}},
		{TestPair(200, 800), resultPair{min: 200, max: 800}},
		{TestPair(defMax+100, 600), resultPair{min: defMin, max: 600}},
		{TestPair(defMin-100, 600), resultPair{min: defMin, max: 600}},
		{TestPair(300, defMax+100), resultPair{min: 300, max: defMax}},
		{TestPair(300, defMin-100), resultPair{min: 300, max: defMax}},
		{TestPair(defMax+100, defMin-100), resultPair{min: defMin, max: defMax}},
		{TestPair(defMin-100, defMax+100), resultPair{min: defMin, max: defMax}},
	}

	for _, test := range suites {
		exist := ResultPair(rangeMinMax(test.exist.defMin, test.exist.defMax, test.exist.minMaximum...))

		if exist != test.expected {
			t.Error("failed to sort min and max. Got ", exist, ", but expected is ", test.expected)
		}
	}
}

func TestRandomRange(t *testing.T) {
	random := RandomRange(100, 200, 120, 180)

	exist := map[int]bool{}
	expectedLen := 10

	for i := 0; i < expectedLen; i++ {
		exist[random.Int()] = true
	}

	if existLen := len(exist); existLen != expectedLen {
		t.Error("generated ", existLen, " random Int, but expected is ", expectedLen)
	}
}

func TestIsPortAvailable(t *testing.T) {
	random := RandomRange(32000, 64000)

	port := random.Int()
	exist := func() bool {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			t.Error("failed to listen on port #", port, "to test it with ", err)
			return false
		}
		defer ln.Close()

		return isPortAvailable(port)
	}()

	if exist {
		t.Error("failed to check an already listened port. Got ", exist, ", but expected is false")
	}

	port = random.Int()
	if !isPortAvailable(port) {
		t.Error("failed to check a free port. Got ", exist, ", but expected is true")
	}
}

func TestCheckLocalPort(t *testing.T) {
	random := RandomRange(32000, 64000)

	port := random.Int()
	exist := func() int {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			t.Error("failed to start listen on port #", port, "to test it with ", err)
			return 0
		}
		defer ln.Close()

		return CheckLocalPort(port)
	}()

	if exist == 0 || exist == port {
		t.Error("failed to to check an already listened port. Got a port #", exist, ", but expected is not zero and other than ", port)
	}

	port = random.Int()
	exist = CheckLocalPort(port)

	if exist != port {
		t.Error("failed to check a free port. Got a port #", exist, ", but expected is a port #", port)
	}
}

func TestDeadLineTimer(t *testing.T) {
	expected := 100 * time.Millisecond

	start := time.Now()
	<-DeadLineTimer(time.Now().Add(expected), true).C
	exist := time.Since(start)

	if exist <= expected {
		t.Error("failed to catch expeted time. Got", exist, ", but expected", expected)
	}

	mode := func() int {
		select {
		case <-DeadLineTimer(time.Now().Add(expected), false).C:
			return 1
		case <-time.NewTimer(150 * time.Millisecond).C:
			return 2
		}

		return 0
	}()

	if mode != 2 {
		t.Error("failed to create inifinite timer on failure")
	}

	mode = func() int {
		select {
		case <-DeadLineTimer(time.Now().Add(-expected), true).C:
			return 1
		case <-time.NewTimer(150 * time.Millisecond).C:
			return 2
		}

		return 0
	}()

	if mode != 1 {
		t.Error("failed to create zero timer on past time")
	}
}
