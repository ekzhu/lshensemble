package lshensemble

import "errors"

var (
	errDomainSizeOrder = errors.New("Domain records must be sorted in ascending order of size")
)

func bootstrapOptimalPartitions(domains <-chan *DomainRecord, numPart int) []Partition {
	sizes, counts := computeSizeDistribution(domains)
	partitions := optimalPartitions(sizes, counts, numPart)
	return partitions
}

func bootstrapOptimal(index *LshEnsemble, sortedDomains <-chan *DomainRecord) error {
	var currPart int
	var currSize int
	for rec := range sortedDomains {
		if currSize > rec.Size {
			return errDomainSizeOrder
		}
		currSize = rec.Size
		if currSize > index.Partitions[currPart].Upper {
			currPart++
		}
		if currPart >= len(index.Partitions) ||
			!(index.Partitions[currPart].Lower <= currSize &&
				currSize <= index.Partitions[currPart].Upper) {
			return errors.New("Domain records does not match the existing partitions")
		}
		index.Add(rec.Key, rec.Signature, currPart)
	}
	index.Index()
	return nil
}

// BootstrapLshEnsembleOptimal builds an index from domains using optimal
// partitioning.
// The returned index consists of MinHash LSH implemented using LshForest.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash
// functions per "band".
// sortedDomainFactory is factory function that returns a DomainRecord channel
// emitting domains in sorted order by their sizes.
func BootstrapLshEnsembleOptimal(numPart, numHash, maxK int,
	sortedDomainFactory func() <-chan *DomainRecord) (*LshEnsemble, error) {
	partitions := bootstrapOptimalPartitions(sortedDomainFactory(), numPart)
	index := NewLshEnsemble(partitions, numHash, maxK)
	err := bootstrapOptimal(index, sortedDomainFactory())
	if err != nil {
		return nil, err
	}
	return index, nil
}

// BootstrapLshEnsemblePlusOptimal builds an index from domains using optimal
// partitioning.
// The returned index consists of MinHash LSH implemented using LshForestArray.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash
// functions per "band".
// sortedDomainFactory is factory function that returns a DomainRecord channel
// emitting domains in sorted order by their sizes.
func BootstrapLshEnsemblePlusOptimal(numPart, numHash, maxK int,
	sortedDomainFactory func() <-chan *DomainRecord) (*LshEnsemble, error) {
	partitions := bootstrapOptimalPartitions(sortedDomainFactory(), numPart)
	index := NewLshEnsemblePlus(partitions, numHash, maxK)
	err := bootstrapOptimal(index, sortedDomainFactory())
	if err != nil {
		return nil, err
	}
	return index, nil
}

func bootstrapEquiDepth(index *LshEnsemble, totalNumDomains int, sortedDomains <-chan *DomainRecord) error {
	numPart := len(index.Partitions)
	depth := totalNumDomains / numPart
	var currDepth, currPart int
	var currSize int
	for rec := range sortedDomains {
		if currSize > rec.Size {
			return errDomainSizeOrder
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

// BootstrapLshEnsembleEquiDepth builds an index from a channel of domains
// using equi-depth partitions -- partitions have approximately the same
// number of domains.
// The returned index consists of MinHash LSH implemented using LshForest.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band".
// sortedDomains is a DomainRecord channel emitting domains in sorted order by their sizes.
func BootstrapLshEnsembleEquiDepth(numPart, numHash, maxK, totalNumDomains int,
	sortedDomains <-chan *DomainRecord) (*LshEnsemble, error) {
	index := NewLshEnsemble(make([]Partition, numPart), numHash, maxK)
	err := bootstrapEquiDepth(index, totalNumDomains, sortedDomains)
	if err != nil {
		return nil, err
	}
	return index, nil
}

// BootstrapLshEnsemblePlusEquiDepth builds an index from a channel of domains
// using equi-depth partitions -- partitions have approximately the same
// number of domains.
// The returned index consists of MinHash LSH implemented using LshForestArray.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band".
// sortedDomains is a DomainRecord channel emitting domains in sorted order by their sizes.
func BootstrapLshEnsemblePlusEquiDepth(numPart, numHash, maxK,
	totalNumDomains int, sortedDomains <-chan *DomainRecord) (*LshEnsemble, error) {
	index := NewLshEnsemblePlus(make([]Partition, numPart), numHash, maxK)
	err := bootstrapEquiDepth(index, totalNumDomains, sortedDomains)
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
