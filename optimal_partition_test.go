package lshensemble

import "testing"

func Test_OptimalPartitions(t *testing.T) {
	sizes := make([]int, 100)
	counts := make([]int, 100)
	for i := range sizes {
		sizes[i] = i + 1
		counts[i] = 10
	}
	numPart := 4
	partitions := optimalPartitions(sizes, counts, numPart)
	if len(partitions) != numPart {
		t.Fatal("Incorrect number of partitions returned")
	}
	t.Log(partitions)

	// Special cases
	numPart = 101
	partitions = optimalPartitions(sizes, counts, numPart)
	if len(partitions) != len(sizes) {
		t.Fatal("Number of partitions should be the same as numebr of sizes " +
			"when it is greater or equal to the number of sizes")
	}

	numPart = 1
	partitions = optimalPartitions(sizes, counts, numPart)
	if (len(partitions) != 1) ||
		(partitions[0].Lower != sizes[0] ||
			partitions[0].Upper != sizes[len(sizes)-1]) {
		t.Fatal("numPart = 1 produced incorrect partition.")
	}
}
