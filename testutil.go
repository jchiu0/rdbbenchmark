package rdbbenchmark

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/jchiu0/rdbbenchmark/rdb"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const (
	numKeys           = 1024
	itemSize          = 8
	prefixSameAsStart = true
	queryKeyID        = int(numKeys) / 2
)

func getKey(i int) string {
	return fmt.Sprintf("%08d", i)
}

func getOptions() *rdb.Options {
	opt := rdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	// Assume 8 bytes for prefix. Should be consistent with GetKey.
	opt.SetPrefixExtractor(rdb.NewFixedPrefixTransform(8))
	return opt
}

func getReadOptions() *rdb.ReadOptions {
	ropt := rdb.NewDefaultReadOptions()
	ropt.SetTotalOrderSeek(false)
	ropt.SetPrefixSameAsStart(prefixSameAsStart)
	return ropt
}

func getWriteOptions() *rdb.WriteOptions {
	wopt := rdb.NewDefaultWriteOptions()
	wopt.SetSync(false) // We don't need to do synchronous writes.
	return wopt
}

func benchPointQuery(valSize int, b *testing.B) {
	path, err := ioutil.TempDir("", "rdbbenchmark")
	check(err)
	defer os.RemoveAll(path)
	opt := getOptions()
	db, err := rdb.OpenDb(opt, path)
	check(err)
	ropt := getReadOptions()
	wopt := getWriteOptions()
	val := make([]byte, valSize)
	for i := 0; i < numKeys; i++ {
		db.Put(wopt, []byte(getKey(i)), val)
	}

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

func benchRowScan(valSize int, b *testing.B) {
	path, err := ioutil.TempDir("", "rdbbenchmark")
	check(err)
	defer os.RemoveAll(path)
	opt := getOptions()

	db, err := rdb.OpenDb(opt, path)
	check(err)
	ropt := getReadOptions()
	wopt := getWriteOptions()

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
		if count != numItems {
			log.Fatalf("Wrong number of item: %d vs %d", count, numItems)
		}
		iter.Close()
	}
	b.StopTimer()
}
