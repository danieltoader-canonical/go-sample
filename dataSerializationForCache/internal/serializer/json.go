package serializer

import (
	"encoding/json"

	"dataSerializationForCache/internal/invoice"
)

// JSON is a schema-less, human-readable codec based on encoding/json.
type JSON struct{}

// NewJSON creates a JSON codec.
func NewJSON() *JSON { return &JSON{} }

// Name returns the format name.
func (*JSON) Name() string { return FormatJSON }

// Marshal encodes the invoice as JSON.
func (*JSON) Marshal(inv invoice.Invoice) ([]byte, error) { return json.Marshal(inv) }

// Unmarshal decodes JSON into an invoice.
func (*JSON) Unmarshal(data []byte) (invoice.Invoice, error) {
	var inv invoice.Invoice
	err := json.Unmarshal(data, &inv)
	return inv, err
}
