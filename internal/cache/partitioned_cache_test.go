package cache

import (
	"fmt"
	"math"
	"testing"
)

//func GetNewPartitionedCache(partitions, size int) *PartitionedCache[string, string] {
//	return NewPartitionedCache[string, string](
//		WithMapPartition[string, string](size),
//		WithPartitionsNum[string, string](partitions),
//		WithHashFunc[string, string](GenericHashFunc[string, string]),
//	)
//}

func TestPartitionedMap(t *testing.T) {
	const (
		partitions = 1000
		size       = 1000
		keysNums   = 5_000_000
	)

	hashFunctions := []struct {
		name      string
		hashFunc  HashFunc[string]
		maxStdDev float64
	}{
		{
			name:      "Generic",
			hashFunc:  GenericHashFunc[string],
			maxStdDev: 5,
		},
		{
			name:      "String",
			hashFunc:  StringHashFunc,
			maxStdDev: 5,
		},
	}

	for _, tCase := range hashFunctions {
		t.Run(fmt.Sprintf("HashFuncStdDeviationOfDestibution.%s", tCase.name), func(t *testing.T) {
			t.Parallel()
			pc := NewPartitionedCache[string, string](
				WithMapPartition[string, string](size),
				WithPartitionsNum[string, string](partitions),
				WithHashFunc[string, string](tCase.hashFunc),
			)

			for i := 0; i < keysNums; i++ {
				kv := fmt.Sprintf("%d", i)
				_ = pc.Put(kv, kv)
			}

			mean := float64(keysNums) / float64(partitions)
			devSum := 0.0

			for _, partition := range pc.GetPartitions() {
				dev := float64(partition.Len()) - mean
				devSum += dev * dev
			}

			stdDev := math.Sqrt(devSum / float64(partitions))
			expectedStdDev := math.Sqrt(mean)
			dev := math.Abs(expectedStdDev - stdDev)

			if dev > tCase.maxStdDev {
				t.Errorf("Standart deviation more than expected. Actual %f; Expected: %f", tCase.maxStdDev, dev)
			}

			fmt.Printf("Standard deviation: %.2f \n", stdDev)
			fmt.Printf("Expected standard deviation: %.2f keys\n", expectedStdDev)
		})
	}
}

func BenchmarkMapSet(b *testing.B) {
	//b.Run("benchmark standard map set", func(b *testing.B) {
	//	var (
	//		wg sync.WaitGroup
	//		mx sync.RWMutex
	//	)
	//
	//	m := make(map[string]string)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			mx.Lock()
	//
	//			kv := fmt.Sprintf("%d", index)
	//
	//			m[kv] = kv
	//			mx.Unlock()
	//		}(i)
	//	}
	//
	//	wg.Wait()
	//})
	//
	//b.Run("benchmark sync map set", func(b *testing.B) {
	//	var (
	//		wg sync.WaitGroup
	//		m  sync.Map
	//	)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			kv := fmt.Sprintf("%d", index)
	//
	//			m.Store(kv, kv)
	//		}(i)
	//	}
	//
	//	wg.Wait()
	//})
	//
	//b.Run("benchmark partitioned map set", func(b *testing.B) {
	//	const (
	//		partitions = 10
	//		size       = 0
	//	)
	//
	//	var wg sync.WaitGroup
	//
	//	pm := GetNewPartitionedCache(partitions, size)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			kv := fmt.Sprintf("%d", index)
	//			_ = pm.Put(kv, kv)
	//		}(i)
	//	}
	//
	//	wg.Wait()
	//})
}

func BenchmarkMapGet(b *testing.B) {
	//b.Run("benchmark standard map get", func(b *testing.B) {
	//	var (
	//		wg sync.WaitGroup
	//		mx sync.RWMutex
	//	)
	//	m := make(map[string]string)
	//	for i := 0; i < b.N; i++ {
	//		kv := fmt.Sprintf("%d", i)
	//
	//		m[kv] = kv
	//	}
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			mx.RLock()
	//			kv := fmt.Sprintf("%d", index)
	//
	//			v, ok := m[kv]
	//			mx.RUnlock()
	//			if !ok && v == "" {
	//				b.Fail()
	//			}
	//		}(i)
	//	}
	//
	//	wg.Wait()
	//})
	//
	//b.Run("benchmark sync map get", func(b *testing.B) {
	//	var (
	//		wg sync.WaitGroup
	//		m  sync.Map
	//	)
	//	for i := 0; i < b.N; i++ {
	//		kv := fmt.Sprintf("%d", i)
	//		m.Store(kv, kv)
	//	}
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			kv := fmt.Sprintf("%d", index)
	//			v, ok := m.Load(kv)
	//			if !ok || v == "" {
	//				b.Fail()
	//			}
	//		}(i)
	//	}
	//
	//	wg.Wait()
	//})
	//
	//b.Run("benchmark partitioned map get", func(b *testing.B) {
	//	const (
	//		partitions = 10
	//		size       = 0
	//	)
	//
	//	pm := GetPartitionedMap(partitions, size)
	//
	//	for i := 0; i < b.N; i++ {
	//		kv := fmt.Sprintf("%d", i)
	//		_ = pm.Put(kv, kv)
	//	}
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	var wg sync.WaitGroup
	//
	//	for i := 0; i < b.N; i++ {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			key := fmt.Sprintf("%d", index)
	//
	//			if v, err := pm.Get(key); err != nil || v == "" {
	//				b.Fail()
	//			}
	//		}(i)
	//	}
	//
	//	wg.Wait()
	//})
}

func BenchmarkMapGetSet(b *testing.B) {
	//b.Run("benchmark standard map get set", func(b *testing.B) {
	//	var mx sync.RWMutex
	//	m := make(map[string]string)
	//	c := make(chan int, 0xff)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	go func() {
	//		var wg sync.WaitGroup
	//		for i := 0; i < b.N; i++ {
	//			wg.Add(1)
	//			go func(index int) {
	//				defer wg.Done()
	//				mx.Lock()
	//				kv := fmt.Sprintf("%d", index)
	//				m[kv] = kv
	//				mx.Unlock()
	//				c <- index
	//			}(i)
	//		}
	//		wg.Wait()
	//		close(c)
	//	}()
	//
	//	var wg sync.WaitGroup
	//	for i := range c {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			mx.RLock()
	//			kv := fmt.Sprintf("%d", index)
	//			v, ok := m[kv]
	//			mx.RUnlock()
	//			if !ok || v == "" {
	//				b.Fail()
	//			}
	//		}(i)
	//	}
	//	wg.Wait()
	//})

	//b.Run("benchmark sync map get set", func(b *testing.B) {
	//	var m sync.Map
	//	c := make(chan int, 0xff)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	go func() {
	//		var wg sync.WaitGroup
	//		for i := 0; i < b.N; i++ {
	//			wg.Add(1)
	//			go func(index int) {
	//				defer wg.Done()
	//				m.Store(fmt.Sprintf("%d", index), index)
	//				c <- index
	//			}(i)
	//		}
	//		wg.Wait()
	//		close(c)
	//	}()
	//
	//	var wg sync.WaitGroup
	//	for i := range c {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			v, ok := m.Load(fmt.Sprintf("%d", index))
	//			if !ok && v == 0 {
	//				b.Fail()
	//			}
	//		}(i)
	//	}
	//	wg.Wait()
	//})

	//b.Run("benchmark partitioned map get set", func(b *testing.B) {
	//	const (
	//		partitions = 10
	//		size       = 0
	//	)
	//
	//	pm := GetNewPartitionedCache(partitions, size)
	//
	//	c := make(chan int, 0xff)
	//
	//	b.ReportAllocs()
	//	b.ResetTimer()
	//
	//	go func() {
	//		var wg sync.WaitGroup
	//		for i := 0; i < b.N; i++ {
	//			wg.Add(1)
	//			go func(index int) {
	//				defer wg.Done()
	//				kv := fmt.Sprintf("%d", index)
	//
	//				_ = pm.Put(kv, kv)
	//				c <- index
	//			}(i)
	//		}
	//		wg.Wait()
	//		close(c)
	//	}()
	//
	//	var wg sync.WaitGroup
	//	for i := range c {
	//		wg.Add(1)
	//		go func(index int) {
	//			defer wg.Done()
	//			key := fmt.Sprintf("%d", index)
	//			if v, err := pm.Get(key); err != nil || v == "" {
	//				b.Fail()
	//			}
	//		}(i)
	//	}
	//	wg.Wait()
	//})
}
