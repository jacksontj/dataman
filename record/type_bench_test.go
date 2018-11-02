package record

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

var x interface{}

func BenchmarkRecordGet(b *testing.B) {
	tests := []struct {
		r    Record
		keys [][]string
	}{
		{
			r: Record{
				"a": 1,
				"b": 2,
				"c": map[string]int{"csub": 1},
				"d": map[string]interface{}{"dsub": 1},
			},
			keys: [][]string{
				[]string{"a"},
				[]string{"b"},
				[]string{"c", "csub"},
				[]string{"d", "dsub"},
				[]string{"nonexistent"},
			},
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for _, key := range test.keys {
				b.Run(strings.Join(key, "."), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						x, _ = test.r.Get(key)
					}
				})
			}
		})
	}
}

func BenchmarkRecordSet(b *testing.B) {
	tests := []struct {
		kv map[string]interface{}
	}{
		{
			kv: map[string]interface{}{
				"a":     1,
				"1deep": map[string]interface{}{"1": 1},
				"b.c.d": 1,
			},
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for k, v := range test.kv {
				r := make(Record)
				b.Run(k+"_repeated", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						r.Set(strings.Split(k, "."), v)
					}
				})
				b.Run(k+"_new", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						r := make(Record)
						r.Set(strings.Split(k, "."), v)
					}
				})
			}
		})
	}
}

func BenchmarkRecordFlatten(b *testing.B) {
	tests := []struct {
		r Record
	}{
		{
			r: Record{"a": 1},
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				test.r.Flatten()
			}
		})
	}
}

func BenchmarkRecordProject(b *testing.B) {
	tests := []struct {
		r    Record
		keys [][][]string
	}{
		{
			r: Record{
				"a": 1,
				"b": 2,
				"c": map[string]int{"csub": 1},
				"d": map[string]interface{}{"dsub": 1},
			},
			keys: [][][]string{
				[][]string{
					[]string{"a"},
				},
				[][]string{
					[]string{"c"},
				},
				[][]string{
					[]string{"c", "csub"},
				},
				[][]string{
					[]string{"nonexistent"},
				},
				[][]string{
					[]string{"a"},
					[]string{"b"},
					[]string{"c", "csub"},
					[]string{"d", "dsub"},
					[]string{"nonexistent"},
				},
			},
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for _, fields := range test.keys {
				b.Run(fmt.Sprintf("%v", fields), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						test.r.Project(fields)
					}
				})
			}
		})
	}
}
