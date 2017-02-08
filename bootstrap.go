package lshensemble

import "errors"

var (
	ErrDomainSizeOrder = errors.New("Domain records must be sorted in ascending order of size")
)

func bootstrap(index *LshEnsemble, totalNumDomains int, sortedDomains <-chan *DomainRecord) error {
	numPart := len(index.Partitions)
	depth := totalNumDomains / numPart
	var currDepth, currPart int
	var currSize int
	for rec := range sortedDomains {
		if currSize > rec.Size {
			return ErrDomainSizeOrder
		}
		currSize = rec.Size
		index.Add(rec.Key, rec.Signature, currPart)
		currDepth++
		index.Partitions[currPart].Upper = rec.Size
		if currDepth >= depth && currPart < numPart-1 {
			currPart++
			index.Partitions[currPart].Lower = rec.Size
			currDepth = 0
		}
	}
	index.Index()
	return nil
}

// BoostrapLshEnsemble builds an index from a channel of domains.
// The returned index consists of MinHash LSH implemented using LshForest.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band".
// sortedDomains is a DomainRecord channel emitting domains in sorted order by their sizes.
func BootstrapLshEnsemble(numPart, numHash, maxK, totalNumDomains int, sortedDomains <-chan *DomainRecord) (*LshEnsemble, error) {
	index := NewLshEnsemble(make([]Partition, numPart), numHash, maxK)
	err := bootstrap(index, totalNumDomains, sortedDomains)
	if err != nil {
		return nil, err
	}
	return index, nil
}

// BoostrapLshEnsemblePlus builds an index from a channel of domains.
// The returned index consists of MinHash LSH implemented using LshForestArray.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band".
// sortedDomains is a DomainRecord channel emitting domains in sorted order by their sizes.
func BootstrapLshEnsemblePlus(numPart, numHash, maxK, totalNumDomains int, sortedDomains <-chan *DomainRecord) (*LshEnsemble, error) {
	index := NewLshEnsemblePlus(make([]Partition, numPart), numHash, maxK)
	err := bootstrap(index, totalNumDomains, sortedDomains)
	if err != nil {
		return nil, err
	}
	return index, nil
}

// Recs2Chan is a utility function that converts a DomainRecord slice in memory to a DomainRecord channel.
func Recs2Chan(recs []*DomainRecord) <-chan *DomainRecord {
	c := make(chan *DomainRecord, 1000)
	go func() {
		for _, r := range recs {
			c <- r
		}
		close(c)
	}()
	return c
}
