package partitioned_map

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartitionedMap(t *testing.T) {
	const (
		partitions = 2
		size       = 6
	)

	hashFunc := func(key string) int {
		return len(key)
	}

	pm := NewPartitionedMap[string, string](partitions, size, hashFunc)

	assert.Equal(t, partitions, len(pm.partitions), "actual partitions count is not equal to expected")
	for i := 0; i < partitions; i++ {
		assert.Equal(t, 0, len(pm.partitions[i].mp), fmt.Sprintf("actual len of partition %d is not equal to expected", i))
	}

	pm.Set("a", "a")
	pk := pm.GetPartitionKey("a")

	assert.Equal(t, 1, len(pm.partitions[pk].mp), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))

	pm.Set("ab", "ab")
	pk = pm.GetPartitionKey("ab")

	assert.Equal(t, 1, len(pm.partitions[pk].mp), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))

	pm.Set("abc", "abc")
	pk = pm.GetPartitionKey("abc")

	assert.Equal(t, 2, len(pm.partitions[pk].mp), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))

	pm.Set("abcd", "abc")
	pk = pm.GetPartitionKey("abcd")

	assert.Equal(t, 2, len(pm.partitions[pk].mp), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))
}

func BenchmarkMapSet(b *testing.B) {
	b.Run("benchmark standard map set", func(b *testing.B) {
		var (
			wg sync.WaitGroup
			mx sync.RWMutex
		)

		m := make(map[int]int)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mx.Lock()
				m[index] = index
				mx.Unlock()
			}(i)
		}

		wg.Wait()
	})

	b.Run("benchmark sync map set", func(b *testing.B) {
		var (
			wg sync.WaitGroup
			m  sync.Map
		)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				m.Store(index, index)
			}(i)
		}

		wg.Wait()
	})

	b.Run("benchmark partitioned map set", func(b *testing.B) {
		var wg sync.WaitGroup
		m := NewPartitionedMap[int, int](runtime.NumCPU(), 0, func(key int) int {
			return key
		})
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				m.Set(index, index)
			}(i)
		}

		wg.Wait()
	})
}

func BenchmarkMapGet(b *testing.B) {
	b.Run("benchmark standard map get", func(b *testing.B) {
		var (
			wg sync.WaitGroup
			mx sync.RWMutex
		)
		m := make(map[int]int)
		for i := 0; i < b.N; i++ {
			m[i] = i
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mx.RLock()
				v, ok := m[index]
				mx.RUnlock()
				if !ok && v == 0 {
					b.Fail()
				}
			}(i)
		}

		wg.Wait()
	})

	b.Run("benchmark sync map get", func(b *testing.B) {
		var (
			wg sync.WaitGroup
			m  sync.Map
		)
		for i := 0; i < b.N; i++ {
			m.Store(i, i)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				v, ok := m.Load(index)
				if !ok && v == 0 {
					b.Fail()
				}
			}(i)
		}

		wg.Wait()
	})

	b.Run("benchmark partitioned map get", func(b *testing.B) {
		var wg sync.WaitGroup
		m := NewPartitionedMap[int, int](runtime.NumCPU(), 0, func(key int) int {
			return key
		})
		for i := 0; i < b.N; i++ {
			m.Set(i, i)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				v, err := m.Get(index)
				if err != nil && v == 0 {
					b.Fail()
				}
			}(i)
		}

		wg.Wait()
	})
}

func BenchmarkMapGetSet(b *testing.B) {
	b.Run("benchmark standard map get set", func(b *testing.B) {
		var mx sync.RWMutex
		m := make(map[string]int)
		c := make(chan int, 0xff)

		b.ReportAllocs()
		b.ResetTimer()

		go func() {
			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					mx.Lock()
					m[fmt.Sprintf("%d", index)] = index
					mx.Unlock()
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
				mx.RLock()
				v, ok := m[fmt.Sprintf("%d", index)]
				mx.RUnlock()
				if !ok && v == 0 {
					b.Fail()
				}
			}(i)
		}
		wg.Wait()
	})

	b.Run("benchmark sync map get set", func(b *testing.B) {
		var m sync.Map
		c := make(chan int, 0xff)

		b.ReportAllocs()
		b.ResetTimer()

		go func() {
			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					m.Store(fmt.Sprintf("%d", index), index)
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
				v, ok := m.Load(fmt.Sprintf("%d", index))
				if !ok && v == 0 {
					b.Fail()
				}
			}(i)
		}
		wg.Wait()
	})

	b.Run("benchmark partitioned map get set", func(b *testing.B) {
		m := NewPartitionedMap[int, int](uint64(runtime.NumCPU()), 0, func(key int) int {
			return key
		})

		c := make(chan int, 0xff)

		b.ReportAllocs()
		b.ResetTimer()

		go func() {
			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					m.Set(index, index)
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
				v, err := m.Get(index)
				if err != nil && v == 0 {
					b.Fail()
				}
			}(i)
		}
		wg.Wait()
	})
}
