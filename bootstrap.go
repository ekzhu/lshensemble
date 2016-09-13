package lshensemble

func bootstrap(index *LshEnsemble, totalNumDomains int, sortedDomains chan *Domain) {
	numPart := len(index.Partitions)
	depth := totalNumDomains / numPart
	var currDepth, currPart int
	for rec := range sortedDomains {
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
}

// BoostrapLshEnsemble builds an index from a channel of domains.
// The returned index consists of MinHash LSH implemented using LshForest.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band". 
// sortedDomains is a Domain channel emitting domains in sorted order by their sizes.
func BootstrapLshEnsemble(numPart, numHash, maxK, totalNumDomains int, sortedDomains chan *Domain) *LshEnsemble {
	index := NewLshEnsemble(make([]Partition, numPart), numHash, maxK)
	bootstrap(index, totalNumDomains, sortedDomains)
	return index
}

// BoostrapLshEnsemblePlus builds an index from a channel of domains.
// The returned index consists of MinHash LSH implemented using LshForestArray.
// numPart is the number of partitions to create.
// numHash is the number of hash functions in MinHash.
// maxK is the maximum value for the MinHash parameter K - the number of hash functions per "band". 
// sortedDomains is a Domain channel emitting domains in sorted order by their sizes.
func BootstrapLshEnsemblePlus(numPart, numHash, maxK, totalNumDomains int, sortedDomains chan *Domain) *LshEnsemble {
	index := NewLshEnsemblePlus(make([]Partition, numPart), numHash, maxK)
	bootstrap(index, totalNumDomains, sortedDomains)
	return index
}

// Recs2Chan is a utility function that converts a Domain slice in memory to a Domain channel.
func Recs2Chan(recs []*Domain) chan *Domain {
	c := make(chan *Domain, 1000)
	go func() {
		for _, r := range recs {
			c <- r
		}
		close(c)
	}()
	return c
}
