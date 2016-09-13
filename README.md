# LSH Ensemble

[![Build Status](https://travis-ci.org/ekzhu/lshensemble.svg?branch=master)](https://travis-ci.org/ekzhu/lshensemble)

[Documentation](https://godoc.org/github.com/ekzhu/lshensemble)

Please cite this paper if you use this library in your work:
>Erkang Zhu, Fatemeh Nargesian, Ken Q. Pu, Renée J. Miller:
>LSH Ensemble: Internet-Scale Domain Search. PVLDB 9(12): 1185-1196 (2016)
>[Link](http://www.vldb.org/pvldb/vol9/p1185-zhu.pdf)

[Presentation slides](http://www.cs.toronto.edu/~ekzhu/talks/lshensemble-vldb2016.pdf) @ VLDB 2016, New Delhi.

## Quick Start Guide

Install this library using:

```
go get github.com/ekzhu/lshensemble
```

Import the library in your `import`:

```go
import (
	github.com/ekzhu/lshensemble
)
```

First you need to obtain the domains, and each domain should have a string ID.
For simplicity we represent a domain as `[]string`.
Assuming you have obtained the domains from some dataset,
you can generate the MinHash signatures from the domains.

```go
domains [][]string
keys []string

// ... 
// obtaining domains and keys
// ...

// initializing the domain records to hold the MinHash signatures
domainRecords := make([]*lshensemble.DomainRecord, len(domains))

// set the minhash seed
seed := 42

// set the number of hash functions
numHash := 256

// create the domain records with the signatures
for i := range domains {
	mh := lshensemble.NewMinhash(seed, numHash)
	for _, v := range domains[i] {
		mh.Push([]byte(v))
	}
	domainRecords[i] := &lshensemble.DomainRecord {
		Key       : keys[i],
		Size      : len(domains[i]),
		Signature : mh.Signature()
	}
}
```

Before you can index the domains, you need to sort them in increasing order by
their sizes. `BySize` wrapper allows the domains to tbe sorted using the build-in `sort`
package.

```go
sort.Sort(lshensemble.BySize(domainRecords))
```

Now you can use `BootstrapLshEnsemble` to create an LSH Ensemble index. You need to
specify the number of partitions to use and some other parameters.
The LSH parameter K (number of hash functions per band) is dynamically tuned at query-time,
but the maximum value needs to be specified here.

```go
// set the number of partitions
numPart := 8

// set the maximum value for the MinHash LSH parameter K 
// (number of hash functions per band).
maxK := 4

// create index
index := lshensemble.BootstrapLshEnsemble(numPart, numHash, maxK, len(domainRecords), lshensemble.Recs2Chan(domainRecords))
```

For better memory efficiency when the number of domains is large, 
it's wiser to use Golang channels and goroutines
to pipeline the generation of the signatures, and use disk-based sorting to sort the domain records. 
This is why `BootstrapLshEnsemble` accepts a channel of `*DomainRecord` as input.
For a small number of domains, you simply use `Recs2Chan` to convert the sorted slice of `*DomainRecord`
into a `chan *DomainRecord`.

To help serializing the domain records to disk, you can use `SerializeSignature`
to serialize the signatures.
You need to come up with your own serialization schema for the keys and sizes.

Lastly, you can query the index using `Query` function. The index returns the *candidates*
domains, which may contains false positives - domains that do not meet the containment
threshold. Therefore, you can optionally include a post-processing step to remove
the false positive domains using the original domain values.

```go
// pick a domain as query
querySig := domainRecords[0].Signature
querySize := domainRecords[0].Size

// set the containment threshold
threshold := 0.5

// get the keys of the candidate domains (may contain false positives),
// and the running time. 
results, dur := index.Query(querySig, querySize, threshold)

// ...
// You may want to include a post processing step here to remove 
// false positive domains using the actual domain values.
// ...
```
