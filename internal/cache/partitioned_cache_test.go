package cache

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

func TestPartitionedMap(t *testing.T) {
	const (
		partitions = 1000
		size       = 1000
		keysNums   = 5_000_000
	)

	tt := []struct {
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

	for _, tc := range tt {
		t.Run(fmt.Sprintf("HashFuncStdDeviationOfDestibution.%s", tc.name), func(t *testing.T) {
			t.Parallel()
			pc := NewPartitionedCache[string, string](
				WithMapPartition[string, string](size),
				WithPartitionsNum[string, string](partitions),
				WithHashFunc[string, string](tc.hashFunc),
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

			if dev > tc.maxStdDev {
				t.Errorf("Standart deviation more than expected. Actual %f; Expected: %f", tc.maxStdDev, dev)
			}

			fmt.Printf("Standard deviation: %.2f \n", stdDev)
			fmt.Printf("Expected standard deviation: %.2f keys\n", expectedStdDev)
		})
	}
}

func BenchmarkMapSet(b *testing.B) {
	const (
		partitions = 1000
		size       = 100
	)

	mapCache := NewMapCache[string, string](partitions * size)
	lruCache := NewLRUCache[string, string](partitions * size)
	pc := NewPartitionedCache[string, string](
		WithPartitionsNum[string, string](partitions),
		WithMapPartition[string, string](size),
	)

	bt := []struct {
		name  string
		cache Cache[string, string]
	}{
		{
			name:  "map cache set",
			cache: mapCache,
		},
		{
			name:  "lru cache set",
			cache: lruCache,
		},
		{
			name:  "partitioned map set",
			cache: pc,
		},
	}

	for _, bt := range bt {
		b.Run(bt.name, func(b *testing.B) {
			wg := sync.WaitGroup{}

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()

					kv := fmt.Sprintf("%d", index)

					bt.cache.Put(kv, kv)
				}(i)
			}

			wg.Wait()
		})
	}
}

func BenchmarkMapGet(b *testing.B) {
	const (
		partitions = 1000
		size       = 100
	)

	mapCache := NewMapCache[string, string](partitions * size)
	lruCache := NewLRUCache[string, string](partitions * size)
	pc := NewPartitionedCache[string, string](
		WithPartitionsNum[string, string](partitions),
		WithMapPartition[string, string](size),
	)

	bt := []struct {
		name  string
		cache Cache[string, string]
	}{
		{
			name:  "map cache get",
			cache: mapCache,
		},
		{
			name:  "lru cache get",
			cache: lruCache,
		},
		{
			name:  "partitioned map get",
			cache: pc,
		},
	}

	for _, bt := range bt {
		b.Run(bt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				kv := fmt.Sprintf("%d", i)

				bt.cache.Put(kv, kv)
			}

			b.ReportAllocs()
			b.ResetTimer()

			wg := sync.WaitGroup{}

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					kv := fmt.Sprintf("%d", index)

					v, err := bt.cache.Get(kv)
					if err != nil && v == "" {
						b.Fail()
					}
				}(i)
			}

			wg.Wait()
		})
	}
}

func BenchmarkMapGetSet(b *testing.B) {
	const (
		partitions = 1000
		size       = 100
	)

	mapCache := NewMapCache[string, string](partitions * size)
	lruCache := NewLRUCache[string, string](partitions * size)
	pc := NewPartitionedCache[string, string](
		WithPartitionsNum[string, string](partitions),
		WithMapPartition[string, string](size),
	)

	bt := []struct {
		name  string
		cache Cache[string, string]
	}{
		{
			name:  "map cache set get",
			cache: mapCache,
		},
		{
			name:  "lru cache set get",
			cache: lruCache,
		},
		{
			name:  "partitioned map set get",
			cache: pc,
		},
	}

	for _, bt := range bt {
		b.Run(bt.name, func(b *testing.B) {
			c := make(chan int, 0xff)

			go func() {
				var wg sync.WaitGroup
				for i := 0; i < b.N; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						kv := fmt.Sprintf("%d", index)
						bt.cache.Put(kv, kv)
						c <- index
					}(i)
				}
				wg.Wait()
				close(c)
			}()

			var wg sync.WaitGroup
			for i := range c {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					kv := fmt.Sprintf("%d", index)
					v, err := bt.cache.Get(kv)
					if err != nil || v == "" {
						b.Fail()
					}
				}(i)
			}

			wg.Wait()
		})
	}
}
