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
