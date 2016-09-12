package rdbbenchmark

import "testing"

//func BenchmarkPointQuery7(b *testing.B)  { benchPointQuery(1<<7, b) }
//func BenchmarkPointQuery8(b *testing.B)  { benchPointQuery(1<<8, b) }
//func BenchmarkPointQuery9(b *testing.B)  { benchPointQuery(1<<9, b) }
//func BenchmarkPointQuery10(b *testing.B) { benchPointQuery(1<<10, b) }
//func BenchmarkPointQuery11(b *testing.B) { benchPointQuery(1<<11, b) }
//func BenchmarkPointQuery12(b *testing.B) { benchPointQuery(1<<12, b) }

func BenchmarkRowScan7(b *testing.B)  { benchRowScan(1<<7, b) }
func BenchmarkRowScan8(b *testing.B)  { benchRowScan(1<<8, b) }
func BenchmarkRowScan9(b *testing.B)  { benchRowScan(1<<9, b) }
func BenchmarkRowScan10(b *testing.B) { benchRowScan(1<<10, b) }
func BenchmarkRowScan11(b *testing.B) { benchRowScan(1<<11, b) }
func BenchmarkRowScan12(b *testing.B) { benchRowScan(1<<12, b) }
