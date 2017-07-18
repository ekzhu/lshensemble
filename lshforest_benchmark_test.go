package lshensemble

import (
	"strconv"
	"testing"
)

func Benchmark_LshForest_Insert10000(b *testing.B) {
	sigs := make([][]uint64, 10000)
	for i := range sigs {
		sigs[i] = randomSignature(64, int64(i))
	}
	b.ResetTimer()
	f := NewLshForest16(2, 32)
	for i := range sigs {
		f.Add(strconv.Itoa(i), sigs[i])
	}
	f.Index()
}
