package rdb

// #cgo CXXFLAGS: -std=c++11 -O2
// #cgo LDFLAGS: -lrocksdb -lstdc++
// #include <stdint.h>
// #include <stdlib.h>
// #include "rdbc.h"
import "C"

// A SliceTransform can be used as a prefix extractor.
type SliceTransform interface {
	// Transform a src in domain to a dst in the range.
	Transform(src []byte) []byte

	// Determine whether this is a valid src upon the function applies.
	InDomain(src []byte) bool

	// Determine whether dst=Transform(src) for some src.
	InRange(src []byte) bool

	// Return the name of this transformation.
	Name() string
}

// NewFixedPrefixTransform creates a new fixed prefix transform.
func NewFixedPrefixTransform(prefixLen int) SliceTransform {
	return NewNativeSliceTransform(C.rocksdb_slicetransform_create_fixed_prefix(C.size_t(prefixLen)))
}

// NewNativeSliceTransform creates a SliceTransform object.
func NewNativeSliceTransform(c *C.rocksdb_slicetransform_t) SliceTransform {
	return nativeSliceTransform{c}
}

type nativeSliceTransform struct {
	c *C.rocksdb_slicetransform_t
}

func (st nativeSliceTransform) Transform(src []byte) []byte { return nil }
func (st nativeSliceTransform) InDomain(src []byte) bool    { return false }
func (st nativeSliceTransform) InRange(src []byte) bool     { return false }
func (st nativeSliceTransform) Name() string                { return "" }

// Hold references to slice transforms.
var sliceTransforms []SliceTransform

func registerSliceTransform(st SliceTransform) int {
	sliceTransforms = append(sliceTransforms, st)
	return len(sliceTransforms) - 1
}

//export rdbc_slicetransform_transform
func rdbc_slicetransform_transform(idx int, cKey *C.char, cKeyLen C.size_t, cDstLen *C.size_t) *C.char {
	key := charToByte(cKey, cKeyLen)
	dst := sliceTransforms[idx].Transform(key)
	*cDstLen = C.size_t(len(dst))
	return cByteSlice(dst)
}

//export rdbc_slicetransform_in_domain
func rdbc_slicetransform_in_domain(idx int, cKey *C.char, cKeyLen C.size_t) C.uchar {
	key := charToByte(cKey, cKeyLen)
	inDomain := sliceTransforms[idx].InDomain(key)
	return boolToChar(inDomain)
}

//export rdbc_slicetransform_in_range
func rdbc_slicetransform_in_range(idx int, cKey *C.char, cKeyLen C.size_t) C.uchar {
	key := charToByte(cKey, cKeyLen)
	inRange := sliceTransforms[idx].InRange(key)
	return boolToChar(inRange)
}

//export rdbc_slicetransform_name
func rdbc_slicetransform_name(idx int) *C.char {
	return stringToChar(sliceTransforms[idx].Name())
}
