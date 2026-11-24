package invoice_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dataSerializationForCache/internal/invoice"
)

func TestSample(t *testing.T) {
	inv := invoice.Sample("INV-1")
	require.Len(t, inv.Lines, 40)
	assert.Equal(t, invoice.Line{
		SKU:            "SKU-100",
		Description:    "Cloud hosting (monthly), eu-west region, plan S (project 01)",
		Quantity:       1,
		UnitPriceCents: 4999,
	}, inv.Lines[0])
	assert.Equal(t, "SKU-800", inv.Lines[39].SKU)
	assert.Equal(t, int32(5), inv.Lines[39].Quantity)
	assert.Equal(t, invoice.Sample("INV-1"), inv, "sample must be deterministic")
}

func TestTotalCents(t *testing.T) {
	inv := invoice.Invoice{Lines: []invoice.Line{
		{Quantity: 2, UnitPriceCents: 150},
		{Quantity: 3, UnitPriceCents: 1000},
	}}
	assert.Equal(t, int64(3300), inv.TotalCents())
	assert.Zero(t, invoice.Invoice{}.TotalCents())
}
