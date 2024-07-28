package partitioned_map

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetPartitionedMap(partitions, size int) *PartitionedMap[string, string] {
	return NewPartitionedMap[string, string](
		DefaultStringHashFunc,
		WithDefaultPartition[string, string](size),
		WithPartitionsNum[string, string](partitions),
	)
}

func TestPartitionedMap(t *testing.T) {
	const (
		partitions = 2
		size       = 6
	)

	pm := GetPartitionedMap(partitions, size)

	assert.Equal(t, partitions, pm.Len(), "actual partitions count is not equal to expected")
	for i := 0; i < partitions; i++ {
		assert.Equal(t, 0, pm.partitions[i].Len(), fmt.Sprintf("actual len of partition %d is not equal to expected", i))
	}

	pm.Put("a", "a")
	pk := pm.GetPartitionKey("a")

	assert.Equal(t, 1, pm.GetPartition("a").Len(), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))

	pm.Put("ab", "ab")

	assert.Equal(t, 1, pm.GetPartition("ab").Len(), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))

	pm.Put("abc", "abc")
	pk = pm.GetPartitionKey("abc")

	assert.Equal(t, 2, pm.GetPartition("abc").Len(), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))

	pm.Put("abcd", "abcd")
	pk = pm.GetPartitionKey("abcd")

	assert.Equal(t, 2, pm.GetPartition("abcd").Len(), fmt.Sprintf("actual len of partition %d is not equal to expected", pk))
}

func BenchmarkMapSet(b *testing.B) {
	b.Run("benchmark standard map set", func(b *testing.B) {
		var (
			wg sync.WaitGroup
			mx sync.RWMutex
		)

		m := make(map[string]string)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mx.Lock()

				kv := fmt.Sprintf("%d", index)

				m[kv] = kv
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
				kv := fmt.Sprintf("%d", index)

				m.Store(kv, kv)
			}(i)
		}

		wg.Wait()
	})

	b.Run("benchmark partitioned map set", func(b *testing.B) {
		const (
			partitions = 10
			size       = 0
		)

		var wg sync.WaitGroup

		pm := GetPartitionedMap(partitions, size)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				kv := fmt.Sprintf("%d", index)
				pm.Put(kv, kv)
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
		m := make(map[string]string)
		for i := 0; i < b.N; i++ {
			kv := fmt.Sprintf("%d", i)

			m[kv] = kv
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mx.RLock()
				kv := fmt.Sprintf("%d", index)

				v, ok := m[kv]
				mx.RUnlock()
				if !ok && v == "" {
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
			kv := fmt.Sprintf("%d", i)
			m.Store(kv, kv)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				kv := fmt.Sprintf("%d", index)
				v, ok := m.Load(kv)
				if !ok || v == "" {
					b.Fail()
				}
			}(i)
		}

		wg.Wait()
	})

	b.Run("benchmark partitioned map get", func(b *testing.B) {
		const (
			partitions = 10
			size       = 0
		)

		pm := GetPartitionedMap(partitions, size)

		for i := 0; i < b.N; i++ {
			kv := fmt.Sprintf("%d", i)
			pm.Put(kv, kv)
		}

		b.ReportAllocs()
		b.ResetTimer()

		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				key := fmt.Sprintf("%d", index)

				if v, ok := pm.Get(key); !ok || v == "" {
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
		m := make(map[string]string)
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
					kv := fmt.Sprintf("%d", index)
					m[kv] = kv
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
				kv := fmt.Sprintf("%d", index)
				v, ok := m[kv]
				mx.RUnlock()
				if !ok || v == "" {
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
		const (
			partitions = 10
			size       = 0
		)

		pm := GetPartitionedMap(partitions, size)

		c := make(chan int, 0xff)

		b.ReportAllocs()
		b.ResetTimer()

		go func() {
			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					kv := fmt.Sprintf("%d", index)

					pm.Put(kv, kv)
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
				key := fmt.Sprintf("%d", index)
				if v, ok := pm.Get(key); !ok || v == "" {
					b.Fail()
				}
			}(i)
		}
		wg.Wait()
	})
}
