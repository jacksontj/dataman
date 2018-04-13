package metadata

import (
	"strconv"
	"testing"
)

func TestCollectionKeyspaceFindPartition(t *testing.T) {
	tests := []struct {
		ranges [][]uint64
		cases  [][]uint64
	}{
		{
			ranges: [][]uint64{
				[]uint64{1, 5},
				[]uint64{5, 15},
				[]uint64{15, 0},
			},
			cases: [][]uint64{
				[]uint64{1, 0},
				[]uint64{2, 0},
				[]uint64{3, 0},
				[]uint64{4, 0},
				[]uint64{5, 1},
				[]uint64{6, 1},
				[]uint64{11, 1},
				[]uint64{15, 2},
				[]uint64{16, 2},
				[]uint64{10000, 2},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			partitions := make([]*CollectionKeyspacePartition, len(test.ranges))
			for x, r := range test.ranges {
				partitions[x] = &CollectionKeyspacePartition{StartId: r[0], EndId: r[1]}
			}
			keyspace := CollectionKeyspace{Partitions: partitions}

			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					actualPartition := keyspace.GetKeyspacePartition(c[0])

					if actualPartition != partitions[c[1]] {
						t.Fatalf("Wrong result!")
					}
				})
			}
		})
	}
}
