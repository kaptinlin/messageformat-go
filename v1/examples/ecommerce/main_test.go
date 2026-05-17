package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationServiceMessages(t *testing.T) {
	t.Parallel()

	service, err := NewNotificationService("en")
	require.NoError(t, err)

	cartTests := []struct {
		name   string
		count  int
		gender string
		want   string
	}{
		{name: "empty male cart", count: 0, gender: "male", want: "He has no items in his cart"},
		{name: "single female cart item", count: 1, gender: "female", want: "She has 1 item in her cart"},
		{name: "unknown gender falls back", count: 3, gender: "unknown", want: "They have 3 items in their cart"},
	}
	for _, tc := range cartTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := service.CartMessage(tc.count, tc.gender)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	order, err := service.OrderStatusMessage(Order{
		ID:       "ORD-100",
		Items:    2,
		Total:    49.5,
		Customer: "Ada Lovelace",
		Status:   "shipped",
	})
	require.NoError(t, err)
	assert.Contains(t, order, "Great news Ada Lovelace")
	assert.Contains(t, order, "order #ORD-100")
	assert.Contains(t, order, "2 items")

	inventoryTests := []struct {
		name    string
		product string
		stock   int
		want    string
	}{
		{name: "out of stock", product: "Headphones", stock: 0, want: "Headphones is out of stock"},
		{name: "single item", product: "Mouse", stock: 1, want: "Only 1 Mouse left in stock"},
		{name: "multiple items", product: "Cable", stock: 3, want: "3 Cable items available"},
	}
	for _, tc := range inventoryTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := service.InventoryMessage(tc.product, tc.stock)
			require.NoError(t, err)
			assert.True(t, strings.Contains(got, tc.want), "inventory message %q should contain %q", got, tc.want)
		})
	}
}

func TestEcommerceExampleRuns(t *testing.T) {
	silenceStdout(t)

	main()
}

func silenceStdout(t *testing.T) {
	t.Helper()

	originalStdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)

	os.Stdout = devNull
	t.Cleanup(func() {
		os.Stdout = originalStdout
		require.NoError(t, devNull.Close())
	})
}
