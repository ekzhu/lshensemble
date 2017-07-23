package lshensemble

import (
	"math/rand"
	"testing"
)

func randomSignature(size int, seed int64) []uint64 {
	r := rand.New(rand.NewSource(seed))
	sig := make([]uint64, size)
	for i := range sig {
		sig[i] = uint64(r.Int63())
	}
	return sig
}

func Test_HashKeyFunc16(t *testing.T) {
	sig := randomSignature(2, 1)
	f := hashKeyFuncGen(2)
	hashKey := f(sig)
	if len(hashKey) != 2*2 {
		t.Fatal(len(hashKey))
	}
}

func Test_HashKeyFunc64(t *testing.T) {
	sig := randomSignature(2, 1)
	f := hashKeyFuncGen(8)
	hashKey := f(sig)
	if len(hashKey) != 8*2 {
		t.Fatal(len(hashKey))
	}
}

func Test_LshForest(t *testing.T) {
	f := NewLshForest16(2, 4)
	sig1 := randomSignature(8, 2)
	sig2 := randomSignature(8, 1)
	sig3 := randomSignature(8, 1)

	f.Add("sig1", sig1)
	f.Add("sig2", sig2)
	f.Add("sig3", sig3)
	f.Index()
	for i := range f.hashTables {
		if len(f.hashTables[i]) != 2 {
			t.Fatal(f.hashTables[i])
		}
	}

	keys := make(chan interface{})
	done := make(chan struct{})
	defer close(done)
	go func() {
		f.Query(sig3, 1, 4, keys, done)
		close(keys)
	}()
	found := 0
	for key := range keys {
		if key == "sig2" || key == "sig3" {
			found++
		}
	}
	if found != 2 {
		t.Fatal("unable to retrieve inserted keys")
	}
}

func Test_LshForest_OptimalKL(t *testing.T) {
	f := NewLshForest16(2, 32)
	t.Log(f.OptimalKL(32, 12, 0.5))
}
