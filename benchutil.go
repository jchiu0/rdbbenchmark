package rdbbenchmark

import (
	"fmt"
	"log"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const (
	numKeys        = 1024
	itemSize       = 8
	queryKeyID     = int(numKeys) / 2 // Query for this key.
	blockCacheSize = 256 << 20

	// Everything below should be used only for row scan queries.
	prefixLength      = 8
	prefixSameAsStart = true // If false, slow row scan by >50%.
)

func getKey(i int) string {
	// Should be consistent with prefixLength.
	return fmt.Sprintf("%08d", i)
}
