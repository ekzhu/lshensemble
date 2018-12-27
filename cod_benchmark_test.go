package lshensemble

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	benchmarkSeed = 42
	fracQuery     = 0.01
	minDomainSize = 10
)

var (
	thresholds = []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
)

// Running this function requires a `_cod_domains` directory
// in the current directory.
// The `_code_domains` directory should contain domains files,
// which are line-separated files.
func Benchmark_CanadianOpenData(b *testing.B) {
	// Read raw domains
	start := time.Now()
	rawDomains := make([]rawDomain, 0)
	var count int
	fmt.Println()
	for domain := range readDomains("_cod_domains") {
		// Ignore domaisn with less than 10 values
		if len(domain.values) < minDomainSize {
			continue
		}
		rawDomains = append(rawDomains, domain)
		count++
		fmt.Printf("\rRead %d domains", count)
	}
	fmt.Println()
	log.Printf("Read %d domains in %s", len(rawDomains),
		time.Now().Sub(start).String())

	// Select queries
	numQuery := int(fracQuery * float64(len(rawDomains)))
	queries := make([]rawDomain, 0, numQuery)
	rand.Seed(int64(benchmarkSeed))
	for _, i := range rand.Perm(len(rawDomains))[:numQuery] {
		queries = append(queries, rawDomains[i])
	}

	// Run benchmark
	for _, threshold := range thresholds {
		log.Printf("Canadian Open Data benchmark threshold = %.2f", threshold)
		benchmarkCOD(rawDomains, queries, threshold)
	}
}

func benchmarkCOD(rawDomains, queries []rawDomain, threshold float64) {
	linearscanOutput := fmt.Sprintf("_cod_linearscan_threshold=%.2f", threshold)
	lshensembleOutput := fmt.Sprintf("_cod_lshensemble_threshold=%.2f", threshold)
	accuracyOutput := fmt.Sprintf("_cod_accuracy_threshold=%.2f", threshold)
	benchmarkLinearscan(rawDomains, queries, threshold, linearscanOutput)
	benchmarkLshEnsemble(rawDomains, queries, threshold, lshensembleOutput)
	benchmarkAccuracy(linearscanOutput, lshensembleOutput, accuracyOutput)
}

type rawDomain struct {
	values map[string]bool
	key    string
}

type byKey []*rawDomain

func (ds byKey) Len() int           { return len(ds) }
func (ds byKey) Swap(i, j int)      { ds[i], ds[j] = ds[j], ds[i] }
func (ds byKey) Less(i, j int) bool { return ds[i].key < ds[j].key }

func readDomains(dir string) chan rawDomain {
	out := make(chan rawDomain)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		msg := fmt.Sprintf("Error reading domain directory %s, does it exist?", dir)
		panic(msg)
	}
	go func() {
		for _, file := range files {
			key := file.Name()
			values := make(map[string]bool)
			domainFile, err := os.Open(filepath.Join(dir, key))
			if err != nil {
				panic(err)
			}
			scanner := bufio.NewScanner(domainFile)
			for scanner.Scan() {
				v := strings.ToLower(scanner.Text())
				values[v] = true
				err = scanner.Err()
				if err != nil {
					panic(err)
				}
			}
			domainFile.Close()
			out <- rawDomain{
				values: values,
				key:    key,
			}
		}
		close(out)
	}()
	return out
}

type queryResult struct {
	candidates []interface{}
	queryKey   interface{}
	duration   time.Duration
}

func outputQueryResults(results chan queryResult, outputFilename string) {
	f, err := os.Create(outputFilename)
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(f)
	for result := range results {
		out.WriteString(result.queryKey.(string))
		out.WriteString("\t")
		out.WriteString(result.duration.String())
		out.WriteString("\t")
		for i, candidate := range result.candidates {
			out.WriteString(candidate.(string))
			if i < len(result.candidates)-1 {
				out.WriteString("\t")
			}
		}
		out.WriteString("\n")
	}
	out.Flush()
	f.Close()
}
