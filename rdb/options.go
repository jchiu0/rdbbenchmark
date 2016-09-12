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

// CompressionType specifies the block compression.
// DB contents are stored in a set of blocks, each of which holds a
// sequence of key,value pairs. Each block may be compressed before
// being stored in a file. The following enum describes which
// compression method (if any) is used to compress a block.
type CompressionType uint

// Compression types.
const (
	NoCompression     = CompressionType(C.rocksdb_no_compression)
	SnappyCompression = CompressionType(C.rocksdb_snappy_compression)
	ZLibCompression   = CompressionType(C.rocksdb_zlib_compression)
	Bz2Compression    = CompressionType(C.rocksdb_bz2_compression)
)

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

// SetCompression sets the compression algorithm.
// Default: SnappyCompression, which gives lightweight but fast
// compression.
func (opts *Options) SetCompression(value CompressionType) {
	C.rocksdb_options_set_compression(opts.c, C.int(value))
}

// SetCompressionPerLevel sets different compression algorithm per level.
//
// Different levels can have different compression policies. There
// are cases where most lower levels would like to quick compression
// algorithm while the higher levels (which have more data) use
// compression algorithms that have better compression but could
// be slower. This array should have an entry for
// each level of the database. This array overrides the
// value specified in the previous field 'compression'.
func (opts *Options) SetCompressionPerLevel(value []CompressionType) {
	cLevels := make([]C.int, len(value))
	for i, v := range value {
		cLevels[i] = C.int(v)
	}
	C.rocksdb_options_set_compression_per_level(opts.c, &cLevels[0], C.size_t(len(value)))
}

// SetMinLevelToCompress sets the start level to use compression.
func (opts *Options) SetMinLevelToCompress(value int) {
	C.rocksdb_options_set_min_level_to_compress(opts.c, C.int(value))
}

// SetPlainTableFactory sets a plain table factory with prefix-only seek.
//
// For this factory, you need to set prefix_extractor properly to make it
// work. Look-up will starts with prefix hash lookup for key prefix. Inside the
// hash bucket found, a binary search is executed for hash conflicts. Finally,
// a linear search is used.
//
// keyLen: 			plain table has optimization for fix-sized keys,
// 					which can be specified via keyLen.
// bloomBitsPerKey: the number of bits used for bloom filer per prefix. You
//                  may disable it by passing a zero.
// hashTableRatio:  the desired utilization of the hash table used for prefix
//                  hashing. hashTableRatio = number of prefixes / #buckets
//                  in the hash table
// indexSparseness: inside each prefix, need to build one index record for how
//                  many keys for binary search inside each hash bucket.
// Suggested values:
//   bloomBitsPerKey: 10
//   hashTableRatio: 0.75
//   indexSparseness: 16
//   encodingType: IndexPlain
func (opts *Options) SetPlainTableFactory(keyLen uint32, bloomBitsPerKey int, hashTableRatio float64, indexSparseness int, encodingType int) {
	C.rocksdb_options_set_plain_table_factory(opts.c, C.uint32_t(keyLen), C.int(bloomBitsPerKey), C.double(hashTableRatio), C.size_t(indexSparseness), C.int(encodingType))
}
