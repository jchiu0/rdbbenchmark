// For point queries, we do not need any notion of prefix.
package rdbbenchmark

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/jchiu0/rdbbenchmark/rdb"
)

func pointQueryOptions() *rdb.Options {
	bbto := rdb.NewDefaultBlockBasedTableOptions()
	cache := rdb.NewLRUCache(blockCacheSize)
	bbto.SetBlockCache(cache)

	opt := rdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	opt.SetBlockBasedTableFactory(bbto)
	return opt
}

func pointQueryReadOptions() *rdb.ReadOptions {
	ropt := rdb.NewDefaultReadOptions()
	ropt.SetTotalOrderSeek(false)
	ropt.SetPrefixSameAsStart(false)
	return ropt
}

func pointQueryWriteOptions() *rdb.WriteOptions {
	wopt := rdb.NewDefaultWriteOptions()
	wopt.SetSync(false) // We don't need to do synchronous writes.
	return wopt
}

func pointQueryFlushOptions() *rdb.FlushOptions {
	return rdb.NewDefaultFlushOptions()
}

func benchPointQuery(valSize int, b *testing.B) {
	path, err := ioutil.TempDir("", "rdbbenchmark")
	check(err)
	defer os.RemoveAll(path)
	opt := pointQueryOptions()
	db, err := rdb.OpenDb(opt, path)
	check(err)
	ropt := pointQueryReadOptions()
	wopt := pointQueryWriteOptions()
	fopt := pointQueryFlushOptions()

	val := make([]byte, valSize)
	for i := 0; i < numKeys; i++ {
		db.Put(wopt, []byte(getKey(i)), val)
	}
	db.Flush(fopt)

	queryKey := []byte(getKey(queryKeyID))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice, err := db.Get(ropt, queryKey)
		check(err)
		if slice == nil {
			log.Fatal("Invalid result")
		}
		data := slice.Data()
		if data == nil {
			log.Fatal("Invalid result")
		}
	}
	b.StopTimer()
}
