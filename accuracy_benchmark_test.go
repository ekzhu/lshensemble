package lshensemble

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func benchmark_accuracy(groundTruthFilename, queryResultFilename, outputFilename string) {
	groundTruths := readQueryResultFile(groundTruthFilename)
	queryResults := readQueryResultFile(queryResultFilename)
	precisions := make([]float64, 0)
	recalls := make([]float64, 0)
	for i := range queryResults {
		recall, precision := recallPrecision(queryResults[i], groundTruths[i])
		precisions = append(precisions, precision)
		recalls = append(recalls, recall)
	}
	log.Printf("Mean Precision = %.4f", mean(precisions))
	log.Printf("Mean Recall = %.4f", mean(recalls))

	// Output results
	file, err := os.Create(outputFilename)
	if err != nil {
		panic(err)
	}
	out := csv.NewWriter(file)
	out.Write([]string{"Query", "Precision", "Recall"})
	for i := range queryResults {
		line := []string{
			queryResults[i].queryKey.(string),
			strconv.FormatFloat(precisions[i], 'f', -1, 64),
			strconv.FormatFloat(recalls[i], 'f', -1, 64),
		}
		out.Write(line)
	}
	out.Flush()
	file.Close()
	log.Printf("Accuracy report output to %s", outputFilename)
}

func recallPrecision(result, groundTruth queryResult) (recall, precision float64) {
	if len(groundTruth.candidates) == 0 {
		return 1.0, 1.0
	}
	if len(result.candidates) == 0 {
		return 0.0, 0.0
	}
	truth := make(map[interface{}]bool)
	for _, v := range groundTruth.candidates {
		truth[v] = true
	}
	test := make(map[interface{}]bool)
	for _, v := range result.candidates {
		test[v] = true
	}
	if len(truth) != len(groundTruth.candidates) {
		panic("Ground truth contains duplicates")
	}
	if len(test) != len(result.candidates) {
		panic("Query result contain duplicates!")
	}
	overlap := 0
	for id := range test {
		if _, found := truth[id]; found {
			overlap += 1
		}
	}
	recall = float64(overlap) / float64(len(truth))
	precision = float64(overlap) / float64(len(test))
	return
}

func readQueryResultFile(queryResultFile string) []queryResult {
	results := make([]queryResult, 0)
	file, err := os.Open(queryResultFile)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := strings.Split(scanner.Text(), "\t")
		key := raw[0]
		candidates := make([]interface{}, len(raw[2:]))
		for i := range candidates {
			candidates[i] = raw[2+i]
		}
		dur, err := time.ParseDuration(raw[1])
		if err != nil {
			panic(err)
		}
		results = append(results, queryResult{
			queryKey:   key,
			duration:   dur,
			candidates: candidates,
		})
		err = scanner.Err()
		if err != nil {
			panic(err)
		}
	}
	file.Close()
	return results
}

func mean(a []float64) float64 {
	sum := 0.0
	for _, v := range a {
		sum += v
	}
	return sum / float64(len(a))
}
