package lshensemble

import (
	"sort"
)

type Set struct {
	Data map[string]bool
}

func NewSet() *Set {
	return &Set{
		Data : make(map[string]bool),
	}
}

// Add new element to the set
func (s *Set) Add(t string) {
	s.Data[t] = true
}

// len()
func (s *Set) Len() int {
	return len(s.Data)
}

// DomainRecord represents a domain record.
type DomainRecord struct {
	// The unique key of this domain.
	Key        string
	// The domain size.
	Size int
	// The MinHash signature of this domain.
	Signature  Signature
}

// A wrapper for sorting domains.
type BySize []*DomainRecord

func (rs BySize) Len() int           { return len(rs) }
func (rs BySize) Less(i, j int) bool { return rs[i].Size < rs[j].Size }
func (rs BySize) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// Returns a subset of the domains given the size lower bound and upper bound.
func (rs BySize) Subset(lower, upper int) []*DomainRecord {
	if !sort.IsSorted(rs) {
		panic("Must be sorted by domain size first")
	}
	start, end := -1, -1
	for i := range rs {
		if start == -1 && rs[i].Size >= lower {
			start = i
		}
		if end == -1 && (rs[i].Size > upper || i == len(rs)-1) {
			end = i
			break
		}
	}
	if start == -1 || end == -1 {
		panic("Cannot find such domain size range")
	}
	if end == len(rs)-1 {
		end++
	}
	return []*DomainRecord(rs[start:end])
}
