package lshensemble

import "math"

// Compute the integral of function f, lower limit a, upper limit l, and
// precision defined as the quantize step
func integral(f func(float64) float64, a, b, precision float64) float64 {
	var area float64
	for x := a; x < b; x += precision {
		area += f(x+0.5*precision) * precision
	}
	return area
}

// Probability density function for false positive
func falsePositive(x, q, l, k int) func(float64) float64 {
	return func(t float64) float64 {
		return 1.0 - math.Pow(1.0-math.Pow(t/(1.0+float64(x)/float64(q)-t), float64(k)), float64(l))
	}
}

// Probability density function for false negative
func falseNegative(x, q, l, k int) func(float64) float64 {
	return func(t float64) float64 {
		return 1.0 - (1.0 - math.Pow(1.0-math.Pow(t/(1.0+float64(x)/float64(q)-t), float64(k)), float64(l)))
	}
}

// Compute the cummulative probability of false negative
func probFalseNegative(x, q, l, k int, t, precision float64) float64 {
	fn := falseNegative(x, q, l, k)
	xq := float64(x) / float64(q)
	if xq >= 1.0 {
		return integral(fn, t, 1.0, precision)
	}
	if xq >= t {
		return integral(fn, t, xq, precision)
	} else {
		return 0.0
	}
}

// Compute the cummulative probability of false positive
func probFalsePositive(x, q, l, k int, t, precision float64) float64 {
	fp := falsePositive(x, q, l, k)
	xq := float64(x) / float64(q)
	if xq >= 1.0 {
		return integral(fp, 0.0, t, precision)
	}
	if xq >= t {
		return integral(fp, 0.0, t, precision)
	} else {
		return integral(fp, 0.0, xq, precision)
	}
}
