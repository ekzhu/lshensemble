package lshensemble

import (
	"log"
	"sort"
	"time"
)

const (
	numHash = 256
	numPart = 32
	maxK    = 4
	// useOptimalPartitions = true
	useOptimalPartitions = false
)

func benchmarkLshEnsemble(rawDomains []rawDomain, rawQueries []rawDomain,
	threshold float64, outputFilename string) {
	numHash := 256
	numPart := 32
	maxK := 4

	// Minhash domains
	start := time.Now()
	domainRecords := minhashDomains(rawDomains, numHash)
	log.Printf("Minhash %d domains in %s", len(domainRecords),
		time.Now().Sub(start).String())

	// Minhash queries
	start = time.Now()
	queries := minhashDomains(rawQueries, numHash)
	log.Printf("Minhash %d query domains in %s", len(queries),
		time.Now().Sub(start).String())

	// Start main body of lsh ensemble
	// Indexing
	log.Print("Start building LSH Ensemble index")
	sort.Sort(BySize(domainRecords))
	var index *LshEnsemble
	if useOptimalPartitions {
		index, _ = BootstrapLshEnsemblePlusOptimal(numPart, numHash, maxK,
			func() <-chan *DomainRecord { return Recs2Chan(domainRecords) })
	} else {
		index, _ = BootstrapLshEnsemblePlusEquiDepth(numPart, numHash, maxK,
			len(domainRecords), Recs2Chan(domainRecords))
	}
	log.Print("Finished building LSH Ensemble index")
	// Querying
	log.Printf("Start querying LSH Ensemble index with %d queries", len(queries))
	results := make(chan queryResult)
	go func() {
		for _, query := range queries {
			r, d := index.QueryTimed(query.Signature, query.Size, threshold)
			results <- queryResult{
				queryKey:   query.Key,
				duration:   d,
				candidates: r,
			}
		}
		close(results)
	}()
	outputQueryResults(results, outputFilename)
	log.Printf("Finished querying LSH Ensemble index, output %s", outputFilename)
}

func minhashDomains(rawDomains []rawDomain, numHash int) []*DomainRecord {
	domainRecords := make([]*DomainRecord, 0)
	for _, domain := range rawDomains {
		mh := NewMinhash(benchmarkSeed, numHash)
		for v := range domain.values {
			mh.Push([]byte(v))
		}
		domainRecords = append(domainRecords, &DomainRecord{
			Key:       domain.key,
			Size:      len(domain.values),
			Signature: mh.Signature(),
		})
	}
	return domainRecords
}
