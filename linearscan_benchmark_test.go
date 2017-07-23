package lshensemble

import (
	"log"
	"time"
)

func benchmark_linearscan(rawDomains []rawDomain, queries []rawDomain, threshold float64, outputFilename string) {
	log.Printf("Start Linear Scan with %d queries", len(queries))
	results := make(chan queryResult)
	go func() {
		for _, query := range queries {
			start := time.Now()
			r := make([]interface{}, 0)
			for _, domain := range rawDomains {
				c := computeExactContainment(query.values, domain.values)
				if c < threshold {
					continue
				}
				r = append(r, domain.key)
			}
			d := time.Now().Sub(start)
			results <- queryResult{
				queryKey:   query.key,
				duration:   d,
				candidates: r,
			}
		}
		close(results)
	}()
	outputQueryResults(results, outputFilename)
	log.Printf("Finished Linear Scan, output %s", outputFilename)
}

func computeExactContainment(q, d map[string]bool) float64 {
	if len(q) == 0 || len(d) == 0 {
		return 0.0
	}
	var smaller, bigger *(map[string]bool)
	if len(q) < len(d) {
		smaller, bigger = &(q), &(d)
	} else {
		bigger, smaller = &(q), &(d)
	}
	intersection := 0
	for v := range *smaller {
		if _, exist := (*bigger)[v]; exist {
			intersection++
		}
	}
	return float64(intersection) / float64(len(q))
}
