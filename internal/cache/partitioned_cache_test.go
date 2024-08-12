package cache

import (
	"fmt"
	"math"
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
