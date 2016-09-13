# LSH Ensemble

[![Build Status](https://travis-ci.org/ekzhu/lshensemble.svg?branch=master)](https://travis-ci.org/ekzhu/lshensemble)

[Documentation](https://godoc.org/github.com/ekzhu/lshensemble)

Please cite this paper if you use this library in your work:
>Erkang Zhu, Fatemeh Nargesian, Ken Q. Pu, Renée J. Miller:
>LSH Ensemble: Internet-Scale Domain Search. PVLDB 9(12): 1185-1196 (2016)
>[Link](http://www.vldb.org/pvldb/vol9/p1185-zhu.pdf)

[Presentation slides](http://www.cs.toronto.edu/~ekzhu/talks/lshensemble-vldb2016.pdf) @ VLDB 2016, New Delhi.

## Quick Start Guide

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
domainRecords := make([]*Domain, len(domains))

// set the minhash seed
seed := 42

// set the number of hash functions
numHash := 256

// create the domain records with the signatures
for i := range domains {
	mh := NewMinhash(seed, numHash)
	for _, v := range domains[i] {
		mh.Push([]byte(v))
	}
	domainRecords[i] := &Domain {
		Key       : keys[i],
		Size      : len(domains[i]),
		Signature : mh.Signature()
	}
}
```
