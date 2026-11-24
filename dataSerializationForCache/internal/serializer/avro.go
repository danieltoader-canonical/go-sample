package serializer

import (
	"fmt"

	"github.com/confluentinc/confluent-avro-go/v2"
	"github.com/klauspost/compress/zstd"

	"dataSerializationForCache/internal/invoice"
)

// invoiceAvroSchema describes invoice.Invoice. Avro binary data carries no field names,
// so the same schema is required to read the data back.
const invoiceAvroSchema = `{
  "type": "record",
  "name": "Invoice",
  "namespace": "samples.cache",
  "fields": [
    {"name": "id", "type": "string"},
    {"name": "customer_id", "type": "string"},
    {"name": "currency", "type": "string"},
    {"name": "issued_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
    {"name": "paid", "type": "boolean"},
    {"name": "lines", "type": {"type": "array", "items": {
      "type": "record",
      "name": "Line",
      "fields": [
        {"name": "sku", "type": "string"},
        {"name": "description", "type": "string"},
        {"name": "quantity", "type": "int"},
        {"name": "unit_price_cents", "type": "long"}
      ]
    }}}
  ]
}`

// Avro is a schema-based binary codec whose payload is compressed with zstd.
type Avro struct {
	schema  avro.Schema
	encoder *zstd.Encoder
	decoder *zstd.Decoder
}

// NewAvro parses the schema and creates reusable zstd encoder/decoder instances.
// EncodeAll and DecodeAll are safe for concurrent use, so one Avro value can be shared.
func NewAvro() (*Avro, error) {
	schema, err := avro.Parse(invoiceAvroSchema)
	if err != nil {
		return nil, fmt.Errorf("parse avro schema: %w", err)
	}
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("create zstd encoder: %w", err)
	}
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("create zstd decoder: %w", err)
	}
	return &Avro{schema: schema, encoder: enc, decoder: dec}, nil
}

// Name returns the format name.
func (*Avro) Name() string { return FormatAvro }

// Marshal encodes the invoice as Avro binary and compresses it with zstd.
func (a *Avro) Marshal(inv invoice.Invoice) ([]byte, error) {
	raw, err := avro.Marshal(a.schema, inv)
	if err != nil {
		return nil, fmt.Errorf("avro marshal: %w", err)
	}
	return a.encoder.EncodeAll(raw, nil), nil
}

// Unmarshal decompresses the zstd payload and decodes the Avro binary into an invoice.
func (a *Avro) Unmarshal(data []byte) (invoice.Invoice, error) {
	raw, err := a.decoder.DecodeAll(data, nil)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("zstd decompress: %w", err)
	}
	var inv invoice.Invoice
	if err = avro.Unmarshal(a.schema, raw, &inv); err != nil {
		return invoice.Invoice{}, fmt.Errorf("avro unmarshal: %w", err)
	}
	inv.IssuedAt = inv.IssuedAt.UTC()
	return inv, nil
}
