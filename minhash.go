package lshensemble

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"math/rand"

	minwise "github.com/dgryski/go-minhash"
)

// HashValueSize is 8, the number of byte used for each hash value
const HashValueSize = 8

// Minhash represents a MinHash object
type Minhash struct {
	mw *minwise.MinWise
}

// NewMinhash initializes a MinHash object with a seed and the number of
// hash functions.
func NewMinhash(seed int64, numHash int) *Minhash {
	r := rand.New(rand.NewSource(seed))
	b := binary.BigEndian
	b1 := make([]byte, HashValueSize)
	b2 := make([]byte, HashValueSize)
	b.PutUint64(b1, uint64(r.Int63()))
	b.PutUint64(b2, uint64(r.Int63()))
	fnv1 := fnv.New64a()
	fnv2 := fnv.New64a()
	h1 := func(b []byte) uint64 {
		fnv1.Reset()
		fnv1.Write(b1)
		fnv1.Write(b)
		return fnv1.Sum64()
	}
	h2 := func(b []byte) uint64 {
		fnv2.Reset()
		fnv2.Write(b2)
		fnv2.Write(b)
		return fnv2.Sum64()
	}
	return &Minhash{minwise.NewMinWise(h1, h2, numHash)}
}

// Push a new value to the MinHash object.
// The value should be serialized to byte slice.
func (m *Minhash) Push(b []byte) {
	m.mw.Push(b)
}

// Signature exports the MinHash signature.
func (m *Minhash) Signature() []uint64 {
	return m.mw.Signature()
}

// Containment returns the estimated containment of
// |Q \intersect X| / |Q|.
// q and x are the signatures of Q and X respectively.
// If either size is 0, the result is defined to be 0.
func Containment(q, x []uint64, qSize, xSize int) float64 {
	if qSize == 0 || xSize == 0 {
		return 0.0
	}
	var eq int
	for i, hv := range q {
		if x[i] == hv {
			eq++
		}
	}
	jaccard := float64(eq) / float64(len(q))
	c := (float64(xSize)/float64(qSize) + 1.0) * jaccard / (1.0 + jaccard)
	if c > 1.0 {
		return 1.0
	}
	return c
}

// SigToBytes serializes the signature into byte slice
func SigToBytes(sig []uint64) []byte {
	buf := new(bytes.Buffer)
	for _, v := range sig {
		binary.Write(buf, binary.BigEndian, v)
	}
	return buf.Bytes()
}

// BytesToSig converts a byte slice into a signature
func BytesToSig(data []byte) ([]uint64, error) {
	size := len(data) / HashValueSize
	sig := make([]uint64, size)
	buf := bytes.NewReader(data)
	var v uint64
	for i := range sig {
		if err := binary.Read(buf, binary.BigEndian, &v); err != nil {
			return nil, err
		}
		sig[i] = v
	}
	return sig, nil
}
