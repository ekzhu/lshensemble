package lshensemble

import (
	"encoding/binary"
	"sort"
)

type hashKeyFunc func([]uint64) string

func hashKeyFuncGen(hashValueSize int) hashKeyFunc {
	return func(sig []uint64) string {
		s := make([]byte, hashValueSize*len(sig))
		buf := make([]byte, 8)
		for i, v := range sig {
			binary.LittleEndian.PutUint64(buf, v)
			copy(s[i*hashValueSize:(i+1)*hashValueSize], buf[:hashValueSize])
		}
		return string(s)
	}
}

type sizeCount struct {
	size  int
	count int
}

func computeSizeDistribution(domains <-chan *DomainRecord) (sizes, counts []int) {
	m := make(map[int]int)
	for d := range domains {
		if _, exists := m[d.Size]; !exists {
			m[d.Size] = 0
		}
		m[d.Size]++
	}
	sizeCounts := make([]sizeCount, 0, len(m))
	for size := range m {
		sizeCounts = append(sizeCounts, sizeCount{size, m[size]})
	}
	sort.Slice(sizeCounts, func(i, j int) bool {
		return sizeCounts[i].size < sizeCounts[j].size
	})
	sizes, counts = make([]int, len(sizeCounts)), make([]int, len(sizeCounts))
	for i := range sizeCounts {
		sizes[i] = sizeCounts[i].size
		counts[i] = sizeCounts[i].count
	}
	return sizes, counts
}
