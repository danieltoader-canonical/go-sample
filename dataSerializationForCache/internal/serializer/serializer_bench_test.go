package serializer_test

import (
	"testing"

	"dataSerializationForCache/internal/invoice"
)

// Run with `make bench`. bytes/op is reported as a custom metric so size and speed show together.
func BenchmarkMarshal(b *testing.B) {
	codecs := allCodecs(b)
	inv := invoice.Sample("INV-1001")
	for _, codec := range codecs {
		b.Run(codec.Name(), func(b *testing.B) {
			var size int
			for b.Loop() {
				data, err := codec.Marshal(inv)
				if err != nil {
					b.Fatal(err)
				}
				size = len(data)
			}
			b.ReportMetric(float64(size), "payload-bytes")
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	codecs := allCodecs(b)
	inv := invoice.Sample("INV-1001")
	for _, codec := range codecs {
		data, marshalErr := codec.Marshal(inv)
		if marshalErr != nil {
			b.Fatal(marshalErr)
		}
		b.Run(codec.Name(), func(b *testing.B) {
			for b.Loop() {
				if _, err := codec.Unmarshal(data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
