package metadata

import "testing"

var benchFields map[string]*CollectionField

type ValidateBench func(*testing.B, *CollectionField)

var validateFuncs map[string]ValidateBench

// TODO: generalize this so we can use the same looping mechanism in testing to
// get all possible permutations in both testing and benchmarking
func init() {
	validateFuncs = map[string]ValidateBench{
		"string":  benchmark_String,
		"bool":    benchmark_Bool,
		"int":     benchmark_int,
		"int8":    benchmark_int8,
		"int16":   benchmark_int16,
		"int32":   benchmark_int32,
		"int64":   benchmark_int64,
		"uint":    benchmark_uint,
		"uint8":   benchmark_uint8,
		"uint16":  benchmark_uint16,
		"uint32":  benchmark_uint32,
		"uint64":  benchmark_uint64,
		"float32": benchmark_float32,
		"float64": benchmark_float64,

		// Some other types
		"map": benchmark_map,
	}

	benchFields = map[string]*CollectionField{
		"int": &CollectionField{
			Type: "_int",
		},
		"bool": &CollectionField{
			Type: "_bool",
		},
		"string": &CollectionField{
			Type: "_string",
		},
		"document": &CollectionField{
			Type: "_document",
			SubFields: map[string]*CollectionField{
				"name": &CollectionField{
					Name:    "name",
					Type:    "_string",
					NotNull: true,
				},
				"number": &CollectionField{
					Name: "number",
					Type: "_int",
				},
				"subDoc": &CollectionField{
					Type: "_document",
					SubFields: map[string]*CollectionField{
						"name": &CollectionField{
							Name:    "name",
							Type:    "_string",
							NotNull: true,
						},
					},
				},
			},
		},
	}
}

func BenchmarkFieldValidation(b *testing.B) {
	for fieldName, field := range benchFields {
		b.Run(fieldName, func(b *testing.B) {
			for name, bench := range validateFuncs {
				b.Run(name, func(b *testing.B) { bench(b, field) })
			}
		})

	}

}

func benchmark_Bool(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(true)
	}
}

func benchmark_String(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate("something")
	}
}

func benchmark_int(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(int(4))
	}
}

func benchmark_int8(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(int8(4))
	}
}

func benchmark_int16(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(int16(4))
	}
}

func benchmark_int32(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(int32(4))
	}
}

func benchmark_int64(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(int64(4))
	}
}

func benchmark_uint(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(uint(4))
	}
}

func benchmark_uint8(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(uint8(4))
	}
}

func benchmark_uint16(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(uint16(4))
	}
}

func benchmark_uint32(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(uint32(4))
	}
}

func benchmark_uint64(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(uint64(4))
	}
}

func benchmark_float32(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(float32(4))
	}
}

func benchmark_float64(b *testing.B, field *CollectionField) {
	for n := 0; n < b.N; n++ {
		field.Validate(float64(4))
	}
}

func benchmark_map(b *testing.B, field *CollectionField) {
	tmp := map[string]interface{}{
		"foo": "bar",
	}
	for n := 0; n < b.N; n++ {
		field.Validate(tmp)
	}
}
