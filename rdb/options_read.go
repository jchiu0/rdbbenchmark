package rdb

// #cgo CXXFLAGS: -std=c++11 -O2
// #cgo LDFLAGS: -lrocksdb -lstdc++
// #include <stdint.h>
// #include <stdlib.h>
// #include "rdbc.h"
import "C"

// ReadOptions represent all of the available options when reading from a
// database.
type ReadOptions struct {
	c *C.rocksdb_readoptions_t
}

// NewDefaultReadOptions creates a default ReadOptions object.
func NewDefaultReadOptions() *ReadOptions {
	return NewNativeReadOptions(C.rocksdb_readoptions_create())
}

// NewNativeReadOptions creates a ReadOptions object.
func NewNativeReadOptions(c *C.rocksdb_readoptions_t) *ReadOptions {
	return &ReadOptions{c}
}

// Destroy deallocates the ReadOptions object.
func (opts *ReadOptions) Destroy() {
	C.rocksdb_readoptions_destroy(opts.c)
	opts.c = nil
}

// SetFillCache specify whether the "data block"/"index block"/"filter block"
// read for this iteration should be cached in memory?
// Callers may wish to set this field to false for bulk scans.
// Default: true
func (opts *ReadOptions) SetFillCache(value bool) {
	C.rocksdb_readoptions_set_fill_cache(opts.c, boolToChar(value))
}

// Enable a total order seek regardless of index format (e.g. hash index)
// used in the table. Some table format (e.g. plain table) may not support
// this option.
// If true when calling Get(), we also skip prefix bloom when reading from
// block based table. It provides a way to read exisiting data after
// changing implementation of prefix extractor.
// Default: false
func (opts *ReadOptions) SetTotalOrderSeek(value bool) {
	C.rocksdb_readoptions_set_total_order_seek(opts.c, boolToChar(value))
}

// Enforce that the iterator only iterates over the same prefix as the seek.
// This option is effective only for prefix seeks, i.e. prefix_extractor is
// non-null for the column family and total_order_seek is false.  Unlike
// iterate_upper_bound, prefix_same_as_start only works within a prefix
// but in both directions.
// Default: false
func (opts *ReadOptions) SetPrefixSameAsStart(value bool) {
	C.rocksdb_readoptions_set_prefix_same_as_start(opts.c, boolToChar(value))
}
