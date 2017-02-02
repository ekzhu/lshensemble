package lshensemble

import (
	"encoding/binary"
	"io"
	"sort"
)

type Set struct {
	Data map[string]bool
}

func NewSet() *Set {
	return &Set{
		Data: make(map[string]bool),
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
	Key string
	// The domain size.
	Size int
	// The MinHash signature of this domain.
	Signature Signature
}

// Write serializes the domain record and write to the data stream w, which implements
// the io.Writer interface.
// keyFn is used to convert the string key into bytes.
// Write returns the number of bytes wrote and error if encountered any.
func (r *DomainRecord) Write(w io.Writer, keyFn func(string) ([]byte, error)) (int, error) {
	keyBin, err := keyFn(r.Key)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(keyBin)
	if err != nil {
		return n, err
	}
	size := int64(r.Size)
	if err := binary.Write(w, binary.LittleEndian, size); err != nil {
		return n, err
	}
	n += 8
	length := int32(len(r.Signature))
	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
		return n, err
	}
	n += 4
	for i := range r.Signature {
		if err := binary.Write(w, binary.LittleEndian, r.Signature[i]); err != nil {
			return n, err
		}
		n += 8
	}
	return n, nil
}

// Read deserializes the domain record from a data stream reader.
// keySize is the number of bytes of the serialized key.
// keyFn is used to convert the serialized key into string.
// Read returns the number of bytes read and error if encountered any.
func (r *DomainRecord) Read(reader io.Reader, keySize int, keyFn func([]byte) (string, error)) (int, error) {
	keyBin := make([]byte, keySize)
	n, err := reader.Read(keyBin)
	if err != nil {
		return n, err
	}
	key, err := keyFn(keyBin)
	if err != nil {
		return n, err
	}
	r.Key = key
	var size int64
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return n, err
	}
	r.Size = int(size)
	n += 8
	var length int32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return n, err
	}
	n += 4
	r.Signature = make(Signature, int(length))
	for i := range r.Signature {
		if err := binary.Read(reader, binary.LittleEndian, &(r.Signature[i])); err != nil {
			return n, err
		}
		n += 8
	}
	return n, nil
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
