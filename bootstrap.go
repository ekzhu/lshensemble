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

func BootstrapLshEnsemble(numPart, numHash, maxK, totalNumDomains int, sortedDomains chan *Domain) *LshEnsemble {
	index := NewLshEnsemble(make([]Partition, numPart), numHash, maxK)
	bootstrap(index, totalNumDomains, sortedDomains)
	return index
}

func BootstrapLshEnsemblePlus(numPart, numHash, maxK, totalNumDomains int, sortedDomains chan *Domain) *LshEnsemble {
	index := NewLshEnsemblePlus(make([]Partition, numPart), numHash, maxK)
	bootstrap(index, totalNumDomains, sortedDomains)
	return index
}

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
