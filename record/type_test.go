package record

import (
	"reflect"
	"strconv"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		r Record
		o map[string]interface{}
	}{
		// flat map-- don't break it
		{
			r: map[string]interface{}{"a": "b"},
			o: map[string]interface{}{"a": "b"},
		},
		// nested map, flatten it
		{
			r: map[string]interface{}{"a": map[string]interface{}{"b": "c"}},
			o: map[string]interface{}{"a.b": "c"},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := test.r.Flatten()
			if !reflect.DeepEqual(output, test.o) {
				t.Fatalf("%d: Maps don't match\n%v\n%v", i, output, test.o)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type recordTestCase struct {
		k []string    // Key of value to get
		v interface{} // Value to get back
		e bool        // Boolean of whether it exists or not
	}
	tests := []struct {
		r     Record // Base record
		cases []recordTestCase
	}{
		// flat map-- don't break it
		{
			r: map[string]interface{}{"a": "b"},
			cases: []recordTestCase{
				// Basic working example
				{
					k: []string{"a"},
					v: "b",
					e: true,
				},
				// try to address a subrecord that doesn't exist
				{
					k: []string{"a", "b"},
					v: nil,
					e: false,
				},
				// A key that definitely doesn't exist
				{
					k: []string{"b"},
					v: nil,
					e: false,
				},
				// No key given
				{
					k: []string{},
					v: nil,
					e: false,
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					v, ok := test.r.Get(c.k)
					if ok != c.e {
						t.Fatalf("Mismatch in exists expected=%v actual=%v", c.e, ok)
					}
					if !reflect.DeepEqual(v, c.v) {
						t.Fatalf("Mismatch in values expected=%v actual=%v", c.v, v)
					}
				})
			}
		})
	}
}

func TestSet(t *testing.T) {
	type recordTestCase struct {
		k []string    // Key of value
		v interface{} // Value to set
		e bool        // Boolean for whether this should work
	}
	tests := []struct {
		r     Record // Base record
		cases []recordTestCase
	}{
		// flat map-- don't break it
		{
			r: map[string]interface{}{},
			cases: []recordTestCase{
				// Basic working example
				{
					k: []string{"a"},
					v: "c",
					e: true,
				},
				// set a value sub of a string
				{
					k: []string{"a", "b"},
					v: nil,
					e: false,
				},
				// set a value sub of a string
				{
					k: []string{"b", "b"},
					v: nil,
					e: true,
				},
				// No key given
				{
					k: []string{},
					v: nil,
					e: false,
				},
				// set a map
				{
					k: []string{"map"},
					v: map[string]interface{}{},
					e: true,
				},
				// set a value in a map
				{
					k: []string{"map", "value"},
					v: 1,
					e: true,
				},
				// set a map in a map
				{
					k: []string{"map", "innermap"},
					v: map[string]interface{}{},
					e: true,
				},
				// set a map in a new keyspace
				{
					k: []string{"new", "top", "thing", "for", "stuff"},
					v: 1,
					e: true,
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					ok := test.r.Set(c.k, c.v)
					if ok != c.e {
						t.Fatalf("Mismatch in return expected=%v actual=%v", c.e, ok)
					}
				})
			}
		})
	}
}

func TestRemove(t *testing.T) {
	type recordTestCase struct {
		k []string // Key of value
		e bool     // Boolean for whether this should work
	}
	tests := []struct {
		r     Record // Base record
		cases []recordTestCase
	}{
		// flat map-- don't break it
		{
			r: map[string]interface{}{"a": 1, "map": map[string]interface{}{"v": 1, "innermap": map[string]interface{}{}}},
			cases: []recordTestCase{
				// delete a subkey which doesnt exist
				{
					k: []string{"a", "b"},
					e: false,
				},
				// do a real delete
				{
					k: []string{"a"},
					e: true,
				},
				// delete an already deleted item
				{
					k: []string{"a"},
					e: true,
				},
				// Delete with no key
				{
					k: []string{},
					e: false,
				},
				// Delete sub value
				{
					k: []string{"map", "v"},
					e: true,
				},
				// Delete sub record
				{
					k: []string{"map", "innermap"},
					e: true,
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					ok := test.r.Remove(c.k)
					if ok != c.e {
						t.Fatalf("Mismatch in return expected=%v actual=%v", c.e, ok)
					}
				})
			}
		})
	}
}

func TestPop(t *testing.T) {
	type recordTestCase struct {
		k []string    // Key of value to get
		v interface{} // Value to get back
		e bool        // Boolean of whether it exists or not
	}
	tests := []struct {
		r     Record // Base record
		cases []recordTestCase
	}{
		// flat map-- don't break it
		{
			r: map[string]interface{}{"a": 1, "map": map[string]interface{}{"v": 1, "innermap": map[string]interface{}{}}},
			cases: []recordTestCase{
				// Basic working example
				{
					k: []string{"a"},
					v: 1,
					e: true,
				},
				// pop something we already pop'd
				{
					k: []string{"a"},
					e: false,
				},
				// try to address a subrecord that doesn't exist
				{
					k: []string{"a", "b"},
					e: false,
				},
				// A key that definitely doesn't exist
				{
					k: []string{"b"},
					e: false,
				},
				// No key given
				{
					k: []string{},
					v: nil,
					e: false,
				},
				// pop an existing subrecord
				{
					k: []string{"map", "v"},
					v: 1,
					e: true,
				},
				// pop an existing map
				{
					k: []string{"map"},
					v: map[string]interface{}{"innermap": map[string]interface{}{}},
					e: true,
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					v, ok := test.r.Pop(c.k)
					if ok != c.e {
						t.Fatalf("Mismatch in exists expected=%v actual=%v", c.e, ok)
					}
					if !reflect.DeepEqual(v, c.v) {
						t.Fatalf("Mismatch in values expected=%v actual=%v", c.v, v)
					}
				})
			}
		})
	}
}

func TestProjection(t *testing.T) {
	type recordTestCase struct {
		k [][]string // Key of value to get
		v Record     // Value to get back
	}
	tests := []struct {
		r     Record // Base record
		cases []recordTestCase
	}{
		// flat map-- don't break it
		{
			r: map[string]interface{}{"a": 1, "map": map[string]interface{}{"v": 1, "innermap": map[string]interface{}{}}},
			cases: []recordTestCase{
				// Basic working example
				{
					k: [][]string{{"a"}},
					v: Record{"a": 1},
				},
				// entire map
				{
					k: [][]string{{"map"}},
					v: Record{"map": map[string]interface{}{"v": 1, "innermap": map[string]interface{}{}}},
				},
				// inner value
				{
					k: [][]string{{"map", "v"}},
					v: Record{"map": map[string]interface{}{"v": 1}},
				},
				// inner map
				{
					k: [][]string{{"map", "innermap"}},
					v: Record{"map": map[string]interface{}{"innermap": map[string]interface{}{}}},
				},
				// a and innermap
				{
					k: [][]string{{"a"}, {"map", "innermap"}},
					v: Record{"a": 1, "map": map[string]interface{}{"innermap": map[string]interface{}{}}},
				},
				// No key given
				{
					k: [][]string{{}},
					v: Record{},
				},
				// key that doesn't exist
				{
					k: [][]string{{"nokey"}},
					v: Record{},
				},
				// inner map key that doesn't exist
				{
					k: [][]string{{"map", "innermap", "notthere"}},
					v: Record{"map": map[string]interface{}{"innermap": map[string]interface{}{}}},
				},
				// inner map key that doesn't exist
				{
					k: [][]string{{"map", "innermap", "notthere"}, {"a"}},
					v: Record{"a": 1, "map": map[string]interface{}{"innermap": map[string]interface{}{}}},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					v := test.r.Project(c.k)
					if !reflect.DeepEqual(v, c.v) {
						t.Fatalf("Mismatch in values expected=%v actual=%v", c.v, v)
					}
				})
			}
		})
	}
}
