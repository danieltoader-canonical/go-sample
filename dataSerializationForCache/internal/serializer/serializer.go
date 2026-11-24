// Package serializer converts invoices to and from the byte formats stored in the cache.
package serializer

import (
	"fmt"
	"strings"

	"dataSerializationForCache/internal/invoice"
)

// Codec encodes and decodes an invoice. Implementations must be safe for concurrent use.
type Codec interface {
	// Name is the format identifier used in CLI flags and cache keys.
	Name() string
	Marshal(inv invoice.Invoice) ([]byte, error)
	Unmarshal(data []byte) (invoice.Invoice, error)
}

// Format names accepted by ByName.
const (
	FormatJSON     = "json"
	FormatAvro     = "avro"
	FormatProtobuf = "protobuf"
	FormatMsgpack  = "msgpack"
)

// Formats lists every supported format in display order.
func Formats() []string {
	return []string{FormatJSON, FormatAvro, FormatProtobuf, FormatMsgpack}
}

// ByName returns the codec for a format name.
func ByName(name string) (Codec, error) {
	switch strings.ToLower(name) {
	case FormatJSON:
		return NewJSON(), nil
	case FormatAvro:
		return NewAvro()
	case FormatProtobuf:
		return NewProtobuf(), nil
	case FormatMsgpack:
		return NewMsgpack(), nil
	default:
		return nil, fmt.Errorf("unknown format %q, expected one of %s", name, strings.Join(Formats(), ", "))
	}
}

// All returns one codec per supported format, in the order of Formats.
func All() ([]Codec, error) {
	codecs := make([]Codec, 0, len(Formats()))
	for _, name := range Formats() {
		c, err := ByName(name)
		if err != nil {
			return nil, err
		}
		codecs = append(codecs, c)
	}
	return codecs, nil
}
