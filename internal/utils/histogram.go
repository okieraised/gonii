package utils

import (
	"fmt"
	"math"
)

// Histogram holds a count of values partitioned over buckets.
type Histogram struct {
	// Min is the size of the smallest bucket.
	Min int
	// Max is the size of the biggest bucket.
	Max int
	// Count is the total size of all buckets.
	Count int
	// Buckets over which values are partitioned.
	Buckets []Bucket
}

// Bucket counts a partition of values.
type Bucket struct {
	// Count is the number of values represented in the bucket.
	Count int
	// Min is the low, inclusive bound of the bucket.
	Min float64
	// Max is the high, exclusive bound of the bucket. If
	// this bucket is the last bucket, the bound is inclusive
	// and contains the max value of the histogram.
	Max float64
}

// Hist creates a histogram partitioning input over `bins` buckets.
func Hist(bins int, input []float64) (Histogram, error) {
	if len(input) == 0 || bins == 0 {
		return Histogram{}, nil
	}

	min, max := input[0], input[0]
	for _, val := range input {
		min = math.Min(min, val)
		max = math.Max(max, val)
	}

	if min == max {
		return Histogram{
			Min:     len(input),
			Max:     len(input),
			Count:   len(input),
			Buckets: []Bucket{{Count: len(input), Min: min, Max: max}},
		}, nil
	}

	scale := (max - min) / float64(bins)
	buckets := make([]Bucket, bins)
	for i := range buckets {
		bmin, bmax := float64(i)*scale+min, float64(i+1)*scale+min
		buckets[i] = Bucket{Min: bmin, Max: bmax}
	}

	minC, maxC := 0, 0
	for _, val := range input {
		minx := float64(min)
		xdiff := val - minx
		bi := iMin(int(xdiff/scale), len(buckets)-1)
		if bi < 0 || bi >= len(buckets) {
			return Histogram{}, fmt.Errorf("invalid bi value: %d", bi)
		}
		buckets[bi].Count++
		minC = iMin(minC, buckets[bi].Count)
		maxC = iMax(maxC, buckets[bi].Count)
	}

	return Histogram{
		Min:     minC,
		Max:     maxC,
		Count:   len(input),
		Buckets: buckets,
	}, nil
}

// PowerHist creates a histogram partitioning input over buckets of power `pow`.
func PowerHist(power float64, input []float64) Histogram {
	if len(input) == 0 || power <= 0 {
		return Histogram{}
	}

	minx, maxx := input[0], input[0]
	for _, val := range input {
		minx = math.Min(minx, val)
		maxx = math.Max(maxx, val)
	}

	fromPower := math.Floor(logBase(minx, power))
	toPower := math.Floor(logBase(maxx, power))

	buckets := make([]Bucket, int(toPower-fromPower)+1)
	for i, bkt := range buckets {
		bkt.Min = math.Pow(power, float64(i)+fromPower)
		bkt.Max = math.Pow(power, float64(i+1)+fromPower)
		buckets[i] = bkt
	}

	minC := 0
	maxC := 0
	for _, val := range input {
		powAway := logBase(val, power) - fromPower
		bi := int(math.Floor(powAway))
		buckets[bi].Count++
		minC = iMin(buckets[bi].Count, minC)
		maxC = iMax(buckets[bi].Count, maxC)
	}

	return Histogram{
		Min:     minC,
		Max:     maxC,
		Count:   len(input),
		Buckets: buckets,
	}
}

// Scale gives the scaled count of the bucket at idx, using the provided scale func.
func (h Histogram) Scale(s ScaleFunc, idx int) float64 {
	bkt := h.Buckets[idx]
	scale := s(h.Min, h.Max, bkt.Count)
	return scale
}

// ScaleFunc is the type to implement to scale a histogram.
type ScaleFunc func(min, max, value int) float64

// Linear builds a ScaleFunc that will linearly scale the values of a histogram so that they do not exceed width.
func Linear(width int) ScaleFunc {
	return func(min, max, value int) float64 {
		if min == max {
			return 1
		}
		return float64(value-min) / float64(max-min) * float64(width)
	}
}

func logBase(a, base float64) float64 {
	return math.Log2(a) / math.Log2(base)
}

func iMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func iMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
