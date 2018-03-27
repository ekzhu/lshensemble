package lshensemble

import (
	"fmt"
	"sync"
	"time"

	"github.com/orcaman/concurrent-map"
)

type param struct {
	k int
	l int
}

// Partition represents a domain size partition in the LSH Ensemble index.
type Partition struct {
	Lower int `json:"lower"`
	Upper int `json:"upper"`
}

// Lsh interface is implemented by LshForst and LshForestArray.
type Lsh interface {
	// Add addes a new key into the index, it won't be searchable
	// until the next time Index() is called since the add.
	Add(key interface{}, sig []uint64)
	// Index makes all keys added so far searchable.
	Index()
	// Query searches the index given a minhash signature, and
	// the LSH parameters k and l. Result keys will be written to
	// the channel out.
	// Closing channel done will cancels the query execution.
	Query(sig []uint64, k, l int, out chan<- interface{}, done <-chan struct{})
	// OptimalKL computes the optimal LSH parameters k and l given
	// x, the index domain size, q, the query domain size, and t,
	// the containment threshold. The resulting false positive (fp)
	// and false negative (fn) probabilities are returned as well.
	OptimalKL(x, q int, t float64) (optK, optL int, fp, fn float64)
}

// LshEnsemble represents an LSH Ensemble index.
type LshEnsemble struct {
	Partitions []Partition
	lshes      []Lsh
	maxK       int
	numHash    int
	paramCache cmap.ConcurrentMap
}

// NewLshEnsemble initializes a new index consists of MinHash LSH implemented using LshForest.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band".
func NewLshEnsemble(parts []Partition, numHash, maxK int) *LshEnsemble {
	lshes := make([]Lsh, len(parts))
	for i := range lshes {
		lshes[i] = NewLshForest(maxK, numHash/maxK)
	}
	return &LshEnsemble{
		lshes:      lshes,
		Partitions: parts,
		maxK:       maxK,
		numHash:    numHash,
		paramCache: cmap.New(),
	}
}

// NewLshEnsemblePlus initializes a new index consists of MinHash LSH implemented using LshForestArray.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band".
func NewLshEnsemblePlus(parts []Partition, numHash, maxK int) *LshEnsemble {
	lshes := make([]Lsh, len(parts))
	for i := range lshes {
		lshes[i] = NewLshForestArray(maxK, numHash)
	}
	return &LshEnsemble{
		lshes:      lshes,
		Partitions: parts,
		maxK:       maxK,
		numHash:    numHash,
		paramCache: cmap.New(),
	}
}

// Add a new domain to the index given its partition ID - the index of the partition.
// The added domain won't be searchable until the Index() function is called.
func (e *LshEnsemble) Add(key interface{}, sig []uint64, partInd int) {
	e.lshes[partInd].Add(key, sig)
}

// Index makes all added domains searchable.
func (e *LshEnsemble) Index() {
	for i := range e.lshes {
		e.lshes[i].Index()
	}
}

// Query returns the candidate domain keys in a channel.
// This function is given the MinHash signature of the query domain, sig, the domain size,
// the containment threshold, and a cancellation channel.
// Closing channel done will cancel the query execution.
// The query signature must be generated using the same seed as the signatures of the indexed domains,
// and have the same number of hash functions.
func (e *LshEnsemble) Query(sig []uint64, size int, threshold float64, done <-chan struct{}) <-chan interface{} {
	params := e.computeParams(size, threshold)
	return e.queryWithParam(sig, params, done)
}

// QueryTimed is similar to Query, returns the candidate domain keys in a slice as well as the running time.
func (e *LshEnsemble) QueryTimed(sig []uint64, size int, threshold float64) (result []interface{}, dur time.Duration) {
	// Compute the optimal k and l for each partition
	params := e.computeParams(size, threshold)
	result = make([]interface{}, 0)
	done := make(chan struct{})
	defer close(done)
	start := time.Now()
	for key := range e.queryWithParam(sig, params, done) {
		result = append(result, key)
	}
	dur = time.Since(start)
	return result, dur
}

func (e *LshEnsemble) queryWithParam(sig []uint64, params []param, done <-chan struct{}) <-chan interface{} {
	// Collect candidates from all partitions
	keyChan := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(len(e.lshes))
	for i := range e.lshes {
		go func(lsh Lsh, k, l int) {
			lsh.Query(sig, k, l, keyChan, done)
			wg.Done()
		}(e.lshes[i], params[i].k, params[i].l)
	}
	go func() {
		wg.Wait()
		close(keyChan)
	}()
	return keyChan
}

// Compute the optimal k and l for each partition
func (e *LshEnsemble) computeParams(size int, threshold float64) []param {
	params := make([]param, len(e.Partitions))
	for i, p := range e.Partitions {
		x := p.Upper
		key := cacheKey(x, size, threshold)
		if cached, exist := e.paramCache.Get(key); exist {
			params[i] = cached.(param)
		} else {
			optK, optL, _, _ := e.lshes[i].OptimalKL(x, size, threshold)
			computed := param{optK, optL}
			e.paramCache.Set(key, computed)
			params[i] = computed
		}
	}
	return params
}

// Make a cache key with threshold precision to 2 decimal points
func cacheKey(x, q int, t float64) string {
	return fmt.Sprintf("%.8x %.8x %.2f", x, q, t)
}
