package rdbbenchmark

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/jchiu0/rdbbenchmark/rdb"
)

func rowScanOptions() *rdb.Options {
	bbto := rdb.NewDefaultBlockBasedTableOptions()
	cache := rdb.NewLRUCache(blockCacheSize)
	bbto.SetBlockCache(cache)
	bbto.SetWholeKeyFiltering(false)
	bbto.SetIndexType(rdb.HashSearch)

	// For our test, all queries hit the table, so bloom filters save nothing.
	// For simplicity, there is no need to enable it.
	// bbto.SetFilterPolicy(rdb.NewBloomFilter(10))

	opt := rdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	opt.SetBlockBasedTableFactory(bbto)
	opt.SetPrefixExtractor(rdb.NewFixedPrefixTransform(prefixLength))

	// This speeds up by about ~10%.
	opt.SetHashSkipListRep(10000, 10, 4)

	// Whether there is compression or not does not seem to affect benchmarks
	// because the test is small. However, row scan is still a LOT slower than point
	// queries. This suggests that everything is in block cache during the test,
	// which means it is uncompressed.
	opt.SetCompression(rdb.NoCompression)

	// Changing from block format to plain table format does not speed up row scan
	// in this test at all. This is again because everything is already in memory
	// yet, so there is little gain going to plain table format.
	// opt.SetPlainTableFactory(prefixLength, 10, 0.75, 16, 1)
	return opt
}

func rowScanReadOptions() *rdb.ReadOptions {
	ropt := rdb.NewDefaultReadOptions()
	ropt.SetTotalOrderSeek(false)

	// This speeds things up a lot as we don't have to run ValidWithPrefix which is
	// much slower than valid.
	ropt.SetPrefixSameAsStart(prefixSameAsStart)
	return ropt
}

func rowScanWriteOptions() *rdb.WriteOptions {
	wopt := rdb.NewDefaultWriteOptions()
	wopt.SetSync(false) // We don't need to do synchronous writes.
	return wopt
}

func benchRowScan(valSize int, b *testing.B) {
	path, err := ioutil.TempDir("", "rdbbenchmark")
	check(err)
	defer os.RemoveAll(path)
	opt := rowScanOptions()

	db, err := rdb.OpenDb(opt, path)
	check(err)
	ropt := rowScanReadOptions()
	wopt := rowScanWriteOptions()

	if (valSize % itemSize) != 0 {
		log.Fatalf("Wrong valSize: %d %d", valSize, itemSize)
	}
	numItems := valSize / itemSize
	val := []byte{}
	for i := 0; i < numKeys; i++ {
		for j := 0; j < numItems; j++ {
			db.Put(wopt, []byte(getKey(i)+fmt.Sprintf("%08d", j)), val)
		}
	}

	queryKey := []byte(getKey(queryKeyID))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := db.NewIterator(ropt)
		count := 0
		if prefixSameAsStart {
			for iter.Seek(queryKey); iter.Valid(); iter.Next() {
				count++
			}
		} else {
			for iter.Seek(queryKey); iter.ValidForPrefix(queryKey); iter.Next() {
				count++
			}
		}

		// If we know numbe rof items in advance, and avoid calling Valid(), row scans
		// take about half as much time!
		//		{
		//			iter.Seek(queryKey)
		//			for i := 0; i < numItems; i++ {
		//				iter.Next()
		//			}
		//			count = numItems
		//		}

		if count != numItems {
			log.Fatalf("Wrong number of item: %d vs %d", count, numItems)
		}
		iter.Close()
	}
	b.StopTimer()
}
