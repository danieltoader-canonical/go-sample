// Package invoice holds the domain model that every serializer stores in the cache.
package invoice

import (
	"fmt"
	"time"
)

// Invoice is the domain value cached by the sample. Struct tags cover JSON, MessagePack and Avro;
// Protobuf uses a generated type and explicit mapping (see serializer/protobuf.go).
type Invoice struct {
	ID         string    `json:"id" msgpack:"id" avro:"id"`
	CustomerID string    `json:"customer_id" msgpack:"customer_id" avro:"customer_id"`
	Currency   string    `json:"currency" msgpack:"currency" avro:"currency"`
	IssuedAt   time.Time `json:"issued_at" msgpack:"issued_at" avro:"issued_at"`
	Paid       bool      `json:"paid" msgpack:"paid" avro:"paid"`
	Lines      []Line    `json:"lines" msgpack:"lines" avro:"lines"`
}

// Line is a single billed item. Money is stored in cents to avoid floating point rounding.
type Line struct {
	SKU            string `json:"sku" msgpack:"sku" avro:"sku"`
	Description    string `json:"description" msgpack:"description" avro:"description"`
	Quantity       int32  `json:"quantity" msgpack:"quantity" avro:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents" msgpack:"unit_price_cents" avro:"unit_price_cents"`
}

// TotalCents returns the sum of all line totals.
func (inv Invoice) TotalCents() int64 {
	var total int64
	for _, l := range inv.Lines {
		total += int64(l.Quantity) * l.UnitPriceCents
	}
	return total
}

// sampleLineCount is the number of lines in the sample invoice, enough to make
// size and speed differences between formats visible.
const sampleLineCount = 40

// catalog lists the products the sample invoice lines are built from.
var catalog = []Line{
	{SKU: "SKU-100", Description: "Cloud hosting (monthly), eu-west region, plan S", UnitPriceCents: 4999},
	{SKU: "SKU-200", Description: "Support hours, business tier, response within 4h", UnitPriceCents: 7500},
	{SKU: "SKU-300", Description: "Extra storage 100GB, replicated across 3 zones", UnitPriceCents: 1250},
	{SKU: "SKU-400", Description: "Managed PostgreSQL instance, 2 vCPU / 8GB RAM", UnitPriceCents: 8900},
	{SKU: "SKU-500", Description: "Outbound bandwidth, per 100GB block", UnitPriceCents: 900},
	{SKU: "SKU-600", Description: "Load balancer with TLS termination", UnitPriceCents: 1999},
	{SKU: "SKU-700", Description: "Daily backups, 30-day retention", UnitPriceCents: 650},
	{SKU: "SKU-800", Description: "Monitoring and alerting, per 10 hosts", UnitPriceCents: 1500},
}

// Sample returns a deterministic invoice used by the CLI and tests: 40 lines cycling through the
// catalog, each tagged with a project number. IssuedAt is UTC with millisecond precision so it
// round-trips through every format unchanged.
func Sample(id string) Invoice {
	lines := make([]Line, 0, sampleLineCount)
	for i := range sampleLineCount {
		item := catalog[i%len(catalog)]
		item.Description = fmt.Sprintf("%s (project %02d)", item.Description, i/len(catalog)+1)
		item.Quantity = int32(i%5 + 1)
		lines = append(lines, item)
	}
	return Invoice{
		ID:         id,
		CustomerID: "CUST-42",
		Currency:   "EUR",
		IssuedAt:   time.Date(2026, time.January, 15, 10, 30, 0, 0, time.UTC),
		Lines:      lines,
	}
}
