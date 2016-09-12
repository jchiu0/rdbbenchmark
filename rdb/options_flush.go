package rdb

// #cgo CXXFLAGS: -std=c++11 -O2
// #cgo LDFLAGS: -lrocksdb -lstdc++
// #include <stdint.h>
// #include <stdlib.h>
// #include "rdbc.h"
import "C"

// FlushOptions represent all of the available options when manual flushing the
// database.
type FlushOptions struct {
	c *C.rocksdb_flushoptions_t
}

// NewDefaultFlushOptions creates a default FlushOptions object.
func NewDefaultFlushOptions() *FlushOptions {
	return NewNativeFlushOptions(C.rocksdb_flushoptions_create())
}

// NewNativeFlushOptions creates a FlushOptions object.
func NewNativeFlushOptions(c *C.rocksdb_flushoptions_t) *FlushOptions {
	return &FlushOptions{c}
}
