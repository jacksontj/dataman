package datamantype

import (
	"strconv"
	"testing"
)

func BenchmarkDatamanTypeNormalization(b *testing.B) {
	for DatamanType, valueList := range validValues {
		b.Run(string(DatamanType), func(b *testing.B) {
			for i, val := range valueList {
				b.Run(strconv.Itoa(i), func(b *testing.B) {
					for x := 0; x < b.N; x++ {
						if _, err := DatamanType.Normalize(val); err != nil {
							b.Fatalf("%d DatamanType=%v val=%v err=%s", i, DatamanType, val, err)
						}
					}
				})
			}
		})
	}
}
