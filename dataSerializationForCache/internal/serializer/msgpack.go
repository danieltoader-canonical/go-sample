package serializer

import (
	"github.com/vmihailenco/msgpack/v5"

	"dataSerializationForCache/internal/invoice"
)

// Msgpack is a schema-less binary codec. It keeps field names like JSON but encodes values compactly.
type Msgpack struct{}

// NewMsgpack creates a MessagePack codec.
func NewMsgpack() *Msgpack { return &Msgpack{} }

// Name returns the format name.
func (*Msgpack) Name() string { return FormatMsgpack }

// Marshal encodes the invoice as MessagePack.
func (*Msgpack) Marshal(inv invoice.Invoice) ([]byte, error) { return msgpack.Marshal(inv) }

// Unmarshal decodes MessagePack into an invoice. Times are normalised to UTC because
// msgpack decodes timestamps in the local time zone.
func (*Msgpack) Unmarshal(data []byte) (invoice.Invoice, error) {
	var inv invoice.Invoice
	if err := msgpack.Unmarshal(data, &inv); err != nil {
		return invoice.Invoice{}, err
	}
	inv.IssuedAt = inv.IssuedAt.UTC()
	return inv, nil
}
