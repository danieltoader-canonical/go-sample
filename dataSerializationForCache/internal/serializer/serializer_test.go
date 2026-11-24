package serializer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dataSerializationForCache/internal/invoice"
	"dataSerializationForCache/internal/serializer"
)

// allCodecs returns every codec and fails the test or benchmark if one cannot be built.
func allCodecs(tb testing.TB) []serializer.Codec {
	tb.Helper()
	codecs, err := serializer.All()
	require.NoError(tb, err)
	return codecs
}

func TestCodecsRoundTrip(t *testing.T) {
	codecs := allCodecs(t)

	want := invoice.Sample("INV-1001")
	for _, codec := range codecs {
		t.Run(codec.Name(), func(t *testing.T) {
			data, err := codec.Marshal(want)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			got, err := codec.Unmarshal(data)
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestCodecsRejectGarbage(t *testing.T) {
	codecs := allCodecs(t)

	for _, codec := range codecs {
		t.Run(codec.Name(), func(t *testing.T) {
			_, err := codec.Unmarshal([]byte{0xff, 0xfe, 0xfd})
			assert.Error(t, err)
		})
	}
}

func TestByName(t *testing.T) {
	for _, name := range serializer.Formats() {
		codec, err := serializer.ByName(name)
		require.NoError(t, err)
		assert.Equal(t, name, codec.Name())
	}

	codec, err := serializer.ByName("AVRO")
	require.NoError(t, err)
	assert.Equal(t, serializer.FormatAvro, codec.Name())

	_, err = serializer.ByName("xml")
	assert.ErrorContains(t, err, `unknown format "xml"`)
}

// TestPayloadSizes documents the relative sizes; schema-based binary formats beat field-named ones.
func TestPayloadSizes(t *testing.T) {
	codecs := allCodecs(t)

	sizes := map[string]int{}
	for _, codec := range codecs {
		data, err := codec.Marshal(invoice.Sample("INV-1001"))
		require.NoError(t, err)
		sizes[codec.Name()] = len(data)
		t.Logf("%-8s %4d bytes", codec.Name(), len(data))
	}

	assert.Less(t, sizes[serializer.FormatMsgpack], sizes[serializer.FormatJSON])
	assert.Less(t, sizes[serializer.FormatProtobuf], sizes[serializer.FormatMsgpack])
	assert.Less(t, sizes[serializer.FormatAvro], sizes[serializer.FormatMsgpack])
}
