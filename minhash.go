package lshensemble

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"

	minwise "github.com/dgryski/go-minhash"
)

const HashValueSize = 8

type Minhash struct {
	mw *minwise.MinWise
}

type Signature []uint64

func NewMinhash(seed, size int) *Minhash {
	r := rand.New(rand.NewSource(int64(seed)))
	b := binary.LittleEndian
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
	return &Minhash{minwise.NewMinWise(h1, h2, size)}
}

func (m *Minhash) Push(b []byte) {
	m.mw.Push(b)
}

func (m1 *Minhash) Similarity(m2 *Minhash) float64 {
	return m1.mw.Similarity(m2.mw)
}

func (m *Minhash) Cardinality() int {
	return m.mw.Cardinality()
}

func (m *Minhash) Merge(m2 *Minhash) {
	m.mw.Merge(m2.mw)
}

func (m *Minhash) Signature() Signature {
	return m.mw.Signature()
}

func (sig Signature) Cardinality() int {
	sum := 0.0
	for _, v := range sig {
		sum += -math.Log(float64(math.MaxUint64-v) / float64(math.MaxUint64))
	}
	return int(float64(len(sig)-1) / sum)
}

func (sig Signature) Intersection(sig2 Signature) (count int) {
	for i := range sig2 {
		if sig[i] == sig2[i] {
			count++
		}
	}
	return
}

func (sig Signature) Merge(sig2 Signature) {
	for i, v := range sig2 {
		if v < sig[i] {
			sig[i] = v
		}
	}
}

func (sig Signature) Write(buffer []byte) {
	offset := 0
	for i := range sig {
		binary.LittleEndian.PutUint64(buffer[offset:], sig[i])
		offset += HashValueSize
	}
}

func (sig Signature) Read(buffer []byte) {
	offset := 0
	for i := range sig {
		sig[i] = binary.LittleEndian.Uint64(buffer[offset:])
		offset += HashValueSize
	}
}

func (sig Signature) ByteLen() int {
	return len(sig) * HashValueSize
}

func SerializeSignature(sig Signature) []byte {
	buffer := make([]byte, sig.ByteLen())
	sig.Write(buffer)
	return buffer
}

func DeserializeSignature(buffer []byte) Signature {
	if len(buffer)%HashValueSize != 0 {
		panic("Incorrect length of buffer")
	}
	size := len(buffer) / HashValueSize
	sig := make(Signature, size)
	sig.Read(buffer)
	return sig
}
