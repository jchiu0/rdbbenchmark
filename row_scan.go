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

	opt := rdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	opt.SetBlockBasedTableFactory(bbto)
	opt.SetPrefixExtractor(rdb.NewFixedPrefixTransform(prefixLength))
	//	opt.SetMemtablePrefixBloomBits(100000000)
	opt.SetMemtablePrefixBloomProbes(6)
	opt.SetHashSkipListRep(10000, 10, 4)
	return opt
}

func rowScanReadOptions() *rdb.ReadOptions {
	ropt := rdb.NewDefaultReadOptions()
	ropt.SetTotalOrderSeek(false)
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
		count = numItems
		if count != numItems {
			log.Fatalf("Wrong number of item: %d vs %d", count, numItems)
		}
		iter.Close()
	}
	b.StopTimer()
}
