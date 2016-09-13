package lshensemble

import (
	"fmt"
	"math"
	"testing"
)

func TestMinhash(t *testing.T) {
	m := NewMinhash(1, 256)
	m.Push([]byte("Test some input"))
	sig := m.Signature()
	buf := SerializeSignature(sig)
	sig2 := DeserializeSignature(buf)
	for i, v := range sig {
		if v != sig2[i] {
			t.Fail()
		}
	}
}

func data(size int) [][]byte {
	d := make([][]byte, size)
	for i := range d {
		d[i] = []byte(fmt.Sprintf("salt%d %d", i, size))
	}
	return d
}

func hashing(mh *Minhash, start, end int, data [][]byte) {
	for i := start; i < end; i++ {
		mh.Push(data[i])
	}
}

func benchmark(minhashSize, dataSize int, t *testing.B) {
	if dataSize < 10 {
		fmt.Printf("\n")
		return
	}
	// Data is a set of unique values
	d := data(dataSize)
	// a and b are two subsets of data with some overlaps
	a_start, a_end := 0, int(float64(dataSize)*0.65)
	b_start, b_end := int(float64(dataSize)*0.35), dataSize

	m1 := NewMinhash(1, minhashSize)
	m2 := NewMinhash(1, minhashSize)

	t.ResetTimer()
	hashing(m1, a_start, a_end, d)
	hashing(m2, b_start, b_end, d)

	est := m1.Similarity(m2)
	act := float64(a_end-b_start) / float64(b_end-a_start)
	err := math.Abs(act - est)
	fmt.Printf("Data size: %8d, ", dataSize)
	fmt.Printf("Real resemblance: %.8f, ", act)
	fmt.Printf("Estimated resemblance: %.8f, ", est)
	fmt.Printf("Absolute Error: %.8f\n", err)
}

func BenchmarkMinWise64(b *testing.B) {
	benchmark(64, b.N, b)
}

func BenchmarkMinWise128(b *testing.B) {
	benchmark(128, b.N, b)
}

func BenchmarkMinWise256(b *testing.B) {
	benchmark(256, b.N, b)
}

func BenchmarkMinWise512(b *testing.B) {
	benchmark(512, b.N, b)
}
