package rdb

// #cgo CXXFLAGS: -std=c++11 -O2
// #cgo LDFLAGS: -lrocksdb -lstdc++
// #include <stdint.h>
// #include <stdlib.h>
// #include "rdbc.h"
import "C"

// Options represent all of the available options when opening a database with Open.
type Options struct {
	c *C.rocksdb_options_t

	// Hold references for GC.
	bbto *BlockBasedTableOptions

	// We keep these so we can free their memory in Destroy.
	cst *C.rocksdb_slicetransform_t
}

// NewDefaultOptions creates the default Options.
func NewDefaultOptions() *Options {
	return NewNativeOptions(C.rocksdb_options_create())
}

// NewNativeOptions creates a Options object.
func NewNativeOptions(c *C.rocksdb_options_t) *Options {
	return &Options{c: c}
}

// SetCreateIfMissing specifies whether the database
// should be created if it is missing.
// Default: false
func (opts *Options) SetCreateIfMissing(value bool) {
	C.rocksdb_options_set_create_if_missing(opts.c, boolToChar(value))
}

// SetBlockBasedTableFactory sets the block based table factory.
func (opts *Options) SetBlockBasedTableFactory(value *BlockBasedTableOptions) {
	opts.bbto = value
	C.rocksdb_options_set_block_based_table_factory(opts.c, value.c)
}

// SetPrefixExtractor sets the prefic extractor.
//
// If set, use the specified function to determine the
// prefixes for keys. These prefixes will be placed in the filter.
// Depending on the workload, this can reduce the number of read-IOP
// cost for scans when a prefix is passed via ReadOptions to
// db.NewIterator().
// Default: nil
func (opts *Options) SetPrefixExtractor(value SliceTransform) {
	if nst, ok := value.(nativeSliceTransform); ok {
		opts.cst = nst.c
	} else {
		idx := registerSliceTransform(value)
		opts.cst = C.rdbc_slicetransform_create(C.uintptr_t(idx))
	}
	C.rocksdb_options_set_prefix_extractor(opts.c, opts.cst)
}

// SetMemtablePrefixBloomBits sets the bloom bits for prefix extractor.
//
// If prefix_extractor is set and bloom_bits is not 0, create prefix bloom
// for memtable.
// Default: 0
func (opts *Options) SetMemtablePrefixBloomBits(value uint32) {
	C.rocksdb_options_set_memtable_prefix_bloom_bits(opts.c, C.uint32_t(value))
}

// SetMemtablePrefixBloomProbes sets the number of hash probes per key.
// Default: 6
func (opts *Options) SetMemtablePrefixBloomProbes(value uint32) {
	C.rocksdb_options_set_memtable_prefix_bloom_probes(opts.c, C.uint32_t(value))
}

// SetHashSkipListRep sets a hash skip list as MemTableRep.
//
// It contains a fixed array of buckets, each
// pointing to a skiplist (null if the bucket is empty).
//
// bucketCount:             number of fixed array buckets
// skiplistHeight:          the max height of the skiplist
// skiplistBranchingFactor: probabilistic size ratio between adjacent
//                          link lists in the skiplist
func (opts *Options) SetHashSkipListRep(bucketCount int, skiplistHeight, skiplistBranchingFactor int32) {
	C.rocksdb_options_set_hash_skip_list_rep(opts.c, C.size_t(bucketCount), C.int32_t(skiplistHeight), C.int32_t(skiplistBranchingFactor))
}

// SetHashLinkListRep sets a hashed linked list as MemTableRep.
//
// It contains a fixed array of buckets, each pointing to a sorted single
// linked list (null if the bucket is empty).
//
// bucketCount: number of fixed array buckets
func (opts *Options) SetHashLinkListRep(bucketCount int) {
	C.rocksdb_options_set_hash_link_list_rep(opts.c, C.size_t(bucketCount))
}
