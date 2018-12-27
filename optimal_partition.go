package lshensemble

import (
	"math"
)

// Computes the expected number of false positives caused by using the
// upper bound set size of the set size interval given by indexes l and u.
func computeNFP(l, u int, sizes, counts []int) float64 {
	if l > u {
		panic("l must be less or equal to u")
	}
	var sum float64
	for i := l; i <= u; i++ {
		sum += float64(sizes[u]-sizes[i]) / float64(sizes[u]) * float64(counts[i])
	}
	return sum
}

// Computes the matrix of expected number of false positives for all possible
// sub-intervals of the complete sorted domain of size sizes.
func computeNFPs(sizes, counts []int) [][]float64 {
	nfps := make([][]float64, len(sizes))
	for l := 0; l < len(sizes); l++ {
		nfps[l] = make([]float64, len(sizes))
		for u := l; u < len(sizes); u++ {
			nfps[l][u] = computeNFP(l, u, sizes, counts)
		}
	}
	return nfps
}

// The solution of the sub-problem: total NFPs and the upper bound index of
// the 2nd right-most partition.
type subSolution struct {
	totalNFPs float64
	u1        int
}

// Computes the optimal partitions given the complete domain of sizes and
// computed number of expected false positives for all sub-intervals.
func computeBestPartitions(numPart int, sizes []int, nfps [][]float64) ([]Partition, float64) {
	if numPart < 2 {
		panic("numPart cannot be less than 2")
	}
	if numPart > len(sizes) {
		panic("numPart cannot be greater than number of sizes")
	}
	if numPart == 2 {
		// If the number of partitions is 2, then simply find the upper bound of
		// the first partition, so that the partitioning produces the smallest
		// total expected number of false positives.
		minTotalNFPs := math.MaxFloat64
		var u int
		for u1 := 0; u1 < len(sizes)-1; u1++ {
			totalNFPs := nfps[0][u1] + nfps[u1+1][len(sizes)-1]
			if totalNFPs < minTotalNFPs {
				minTotalNFPs = totalNFPs
				u = u1
			}
		}
		return []Partition{
			Partition{sizes[0], sizes[u]},
			Partition{sizes[u+1], sizes[len(sizes)-1]},
		}, minTotalNFPs
	}
	// Initialize the matrix for storing the sub-problems' solutions.
	// The first axis is the upper bound index of the sub-problem, in which
	// an optimal partitioning of p number of partitions is to be computed.
	// The second axis is the index of p, see p2i below, starting from 2.
	sols := make([][]subSolution, numPart-2)
	for i := range sols {
		sols[i] = make([]subSolution, len(sizes))
	}
	// p is the number of partitions in a sub-problem.
	// p2i translates the number of partitions into the index in the matrix.
	var p2i = func(p int) int { return p - 2 }
	for p := 2; p < numPart; p++ {
		// The possible upper bound indexes of sub problems start from
		// p - 1 which is the smallest index to have p partitions.
		for u := p - 1; u < len(sizes); u++ {
			minTotalNFPs := math.MaxFloat64
			var u1Best int
			if p == 2 {
				for u1 := 0; u1 < u; u1++ {
					totalNFPs := nfps[0][u1] + nfps[u1+1][u]
					if totalNFPs < minTotalNFPs {
						minTotalNFPs = totalNFPs
						u1Best = u1
					}
				}
			} else {
				for u1 := (p - 1) - 1; u1 < u; u1++ {
					totalNFPs := sols[p2i(p-1)][u1].totalNFPs + nfps[u1+1][u]
					if totalNFPs < minTotalNFPs {
						minTotalNFPs = totalNFPs
						u1Best = u1
					}
				}
			}
			sols[p2i(p)][u] = subSolution{minTotalNFPs, u1Best}
		}
	}
	// Initialize partitions.
	partitions := make([]Partition, 0)
	minTotalNFPs := math.MaxFloat64
	// Find where to place the right-most partition -- find the upper bound
	// index of the 2nd right-most partition.
	var u int
	p := numPart
	for u1 := (p - 1) - 1; u1 < len(sizes)-1; u1++ {
		totalNFPs := sols[p2i(p-1)][u1].totalNFPs + nfps[u1+1][len(sizes)-1]
		if totalNFPs < minTotalNFPs {
			u = u1
			minTotalNFPs = totalNFPs
		}
	}
	partitions = append(partitions, Partition{sizes[u+1], sizes[len(sizes)-1]})
	p--
	// Back-track to find the best partitions using the computed results of
	// sub-probelms.
	for p > 1 {
		// For each sub-problem given p and upper bound index u,
		// find the upper bound index (u1) of the 2nd right most partition.
		u1 := sols[p2i(p)][u].u1
		partitions = append(partitions, Partition{sizes[u1+1], sizes[u]})
		// Move on to a smaller sub-problem.
		u = u1
		p--
	}
	// The last partition is the first one.
	partitions = append(partitions, Partition{sizes[0], sizes[u]})
	// Reverse the order so the first comes first.
	for i, j := 0, len(partitions)-1; i < j; i, j = i+1, j-1 {
		partitions[i], partitions[j] = partitions[j], partitions[i]
	}
	return partitions, minTotalNFPs
}

// optimalPartitions takes a set size distribution and number of partitions
// as input and returns the optimal partition boundaries (inclusive) for
// minimizing number of false positives.
func optimalPartitions(sizes, counts []int, numPart int) []Partition {
	if numPart < 2 {
		return []Partition{Partition{sizes[0], sizes[len(sizes)-1]}}
	}
	if numPart >= len(sizes) {
		// If the number of partitions is greater or equal to the complete
		// domain of set sizes, return the perfect partitions.
		partitions := make([]Partition, len(sizes))
		for i := range sizes {
			partitions[i] = Partition{sizes[i], sizes[i]}
		}
		return partitions
	}
	nfps := computeNFPs(sizes, counts)
	partitions, _ := computeBestPartitions(numPart, sizes, nfps)
	return partitions
}
