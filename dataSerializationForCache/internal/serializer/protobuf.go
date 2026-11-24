package serializer

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"dataSerializationForCache/internal/invoice"
	"dataSerializationForCache/internal/serializer/invoicepb"
)

// Protobuf is a schema-based binary codec using code generated from invoicepb/invoice.proto.
// The domain type is mapped to the generated message so the domain stays free of generated code.
type Protobuf struct{}

// NewProtobuf creates a Protobuf codec.
func NewProtobuf() *Protobuf { return &Protobuf{} }

// Name returns the format name.
func (*Protobuf) Name() string { return FormatProtobuf }

// Marshal maps the invoice to the generated message and encodes it.
func (*Protobuf) Marshal(inv invoice.Invoice) ([]byte, error) {
	msg := &invoicepb.Invoice{
		Id:         inv.ID,
		CustomerId: inv.CustomerID,
		Currency:   inv.Currency,
		IssuedAt:   timestamppb.New(inv.IssuedAt),
		Paid:       inv.Paid,
		Lines:      make([]*invoicepb.Line, 0, len(inv.Lines)),
	}
	for _, l := range inv.Lines {
		msg.Lines = append(msg.Lines, &invoicepb.Line{
			Sku:            l.SKU,
			Description:    l.Description,
			Quantity:       l.Quantity,
			UnitPriceCents: l.UnitPriceCents,
		})
	}
	return proto.Marshal(msg)
}

// Unmarshal decodes the message and maps it back to the domain invoice.
func (*Protobuf) Unmarshal(data []byte) (invoice.Invoice, error) {
	var msg invoicepb.Invoice
	if err := proto.Unmarshal(data, &msg); err != nil {
		return invoice.Invoice{}, err
	}
	inv := invoice.Invoice{
		ID:         msg.GetId(),
		CustomerID: msg.GetCustomerId(),
		Currency:   msg.GetCurrency(),
		IssuedAt:   msg.GetIssuedAt().AsTime(),
		Paid:       msg.GetPaid(),
	}
	if n := len(msg.GetLines()); n > 0 {
		inv.Lines = make([]invoice.Line, 0, n)
	}
	for _, l := range msg.GetLines() {
		inv.Lines = append(inv.Lines, invoice.Line{
			SKU:            l.GetSku(),
			Description:    l.GetDescription(),
			Quantity:       l.GetQuantity(),
			UnitPriceCents: l.GetUnitPriceCents(),
		})
	}
	return inv, nil
}
