package database

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/worldline-go/types"
)

var scanValuesBenchmarkSink []any

// This measures destination allocation only, not SQL execution or total transfer
// throughput. Both variants allocate independent destinations for every row.
func BenchmarkScanDestinations(b *testing.B) {
	templates := []any{
		new(any), new(string), new(bool), new(types.Decimal),
		new(types.NullDecimal), new(types.Time), new(types.Null[types.Time]),
		new(types.Null[string]), new(types.Null[bool]),
	}
	for _, columns := range []int{9, 90} {
		valueTypes := make([]any, columns)
		for i := range valueTypes {
			valueTypes[i] = templates[i%len(templates)]
		}
		b.Run(fmt.Sprintf("columns=%d/reflect", columns), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				values := make([]any, len(valueTypes))
				for i, valueType := range valueTypes {
					t := reflect.TypeOf(valueType)
					if t == nil || t.Kind() != reflect.Pointer {
						b.Fatal("invalid scan destination")
					}
					values[i] = reflect.New(t.Elem()).Interface()
				}
				scanValuesBenchmarkSink = values
			}
		})
		b.Run(fmt.Sprintf("columns=%d/typed", columns), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				values, err := newScanValues(valueTypes)
				if err != nil {
					b.Fatal(err)
				}
				scanValuesBenchmarkSink = values
			}
		})
	}
}
