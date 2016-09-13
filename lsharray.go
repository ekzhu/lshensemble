package lshensemble

import (
	"math"
	"sync"
)

type LshForestArray struct {
	maxK    int
	numHash int
	array   []*LshForest
}

func NewLshForestArray(maxK, numHash int) *LshForestArray {
	array := make([]*LshForest, maxK)
	for k := 1; k <= maxK; k++ {
		array[k-1] = NewLshForest(k, numHash/k)
	}
	return &LshForestArray{
		maxK:    maxK,
		numHash: numHash,
		array:   array,
	}
}

func (a *LshForestArray) Add(key string, sig Signature) {
	var wg sync.WaitGroup
	wg.Add(len(a.array))
	for i := range a.array {
		go func(lsh *LshForest) {
			lsh.Add(key, sig)
			wg.Done()
		}(a.array[i])
	}
	wg.Wait()
}

func (a *LshForestArray) Index() {
	var wg sync.WaitGroup
	wg.Add(len(a.array))
	for i := range a.array {
		go func(lsh *LshForest) {
			lsh.Index()
			wg.Done()
		}(a.array[i])
	}
	wg.Wait()
}

func (a *LshForestArray) Query(sig Signature, k, l int, out chan string) {
	a.array[k-1].Query(sig, -1, l, out)
}

// OptimalKL returns the optimal k and l for containment search
// where x is the indexed domain size, q is the query domain size,
// and t is the containment threshold.
func (a *LshForestArray) OptimalKL(x, q int, t float64) (optK, optL int, fp, fn float64) {
	minError := math.MaxFloat64
	for l := 1; l <= a.numHash; l++ {
		for k := 1; k <= a.maxK; k++ {
			if k*l > a.numHash {
				continue
			}
			currFp := probFalsePositive(x, q, l, k, t, integrationPrecision)
			currFn := probFalseNegative(x, q, l, k, t, integrationPrecision)
			currErr := currFn + currFp
			if minError > currErr {
				minError = currErr
				optK = k
				optL = l
				fp = currFp
				fn = currFn
			}
		}
	}
	return
}
