package rdbbenchmark

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/tecbot/gorocksdb"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const (
	numKeys  = 1024
	itemSize = 8
)

func getKey(i int) string {
	return fmt.Sprintf("%08d", i)
}

func benchPointQuery(valSize int, b *testing.B) {
	path, err := ioutil.TempDir("", "rdbbenchmark")
	check(err)
	defer os.RemoveAll(path)
	opt := gorocksdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(opt, path)
	check(err)
	ropt := gorocksdb.NewDefaultReadOptions()
	wopt := gorocksdb.NewDefaultWriteOptions()
	wopt.SetSync(false) // We don't need to do synchronous writes.

	val := make([]byte, valSize)
	for i := 0; i < numKeys; i++ {
		db.Put(wopt, []byte(getKey(i)), val)
	}

	queryKey := []byte(getKey(numKeys / 2))
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

func BenchmarkPointQuery7(b *testing.B)  { benchPointQuery(1<<7, b) }
func BenchmarkPointQuery8(b *testing.B)  { benchPointQuery(1<<8, b) }
func BenchmarkPointQuery9(b *testing.B)  { benchPointQuery(1<<9, b) }
func BenchmarkPointQuery10(b *testing.B) { benchPointQuery(1<<10, b) }
func BenchmarkPointQuery11(b *testing.B) { benchPointQuery(1<<11, b) }
func BenchmarkPointQuery12(b *testing.B) { benchPointQuery(1<<12, b) }

func benchRowScan(valSize int, b *testing.B) {
	path, err := ioutil.TempDir("", "rdbbenchmark")
	check(err)
	defer os.RemoveAll(path)
	opt := gorocksdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(opt, path)
	check(err)
	ropt := gorocksdb.NewDefaultReadOptions()
	wopt := gorocksdb.NewDefaultWriteOptions()
	wopt.SetSync(false) // We don't need to do synchronous writes.

	if (valSize % itemSize) != 0 {
		log.Fatalf("Wrong valSize: %d %d", valSize, itemSize)
	}
	numItems := valSize / itemSize
	val := make([]byte, itemSize)
	for i := 0; i < numKeys; i++ {
		for j := 0; j < numItems; j++ {
			db.Put(wopt, []byte(getKey(i)+fmt.Sprintf("%08d", j)), val)
		}
	}

	queryKey := []byte(getKey(numKeys / 2))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := db.NewIterator(ropt)
		count := 0
		for iter.Seek(queryKey); iter.ValidForPrefix(queryKey); iter.Next() {
			count++
		}
		if count != numItems {
			log.Fatalf("Wrong number of item: %d vs %d", count, numItems)
		}
	}
	b.StopTimer()
}

func BenchmarkRowScan7(b *testing.B)  { benchRowScan(1<<7, b) }
func BenchmarkRowScan8(b *testing.B)  { benchRowScan(1<<8, b) }
func BenchmarkRowScan9(b *testing.B)  { benchRowScan(1<<9, b) }
func BenchmarkRowScan10(b *testing.B) { benchRowScan(1<<10, b) }
func BenchmarkRowScan11(b *testing.B) { benchRowScan(1<<11, b) }
func BenchmarkRowScan12(b *testing.B) { benchRowScan(1<<12, b) }
