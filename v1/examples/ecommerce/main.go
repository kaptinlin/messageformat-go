// Package main demonstrates real-world e-commerce MessageFormat usage.
//
// This example showcases practical MessageFormat applications in an e-commerce context:
//   - Shopping cart status messages with gender selection
//   - Order status notifications with complex nested logic
//   - Inventory alerts with stock level conditions
//   - Service-oriented architecture patterns
//
// Run this example with:
//
//	cd examples/ecommerce && go run main.go
package main

import (
	"fmt"
	"log"

	mf "github.com/kaptinlin/messageformat-go/v1"
)

// Order represents an e-commerce order
type Order struct {
	ID       string
	Items    int
	Total    float64
	Customer string
	Status   string
}

// NotificationService handles user notifications
type NotificationService struct {
	messageFormat *mf.MessageFormat
}

// NewNotificationService creates a new notification service
func NewNotificationService(locale string) (*NotificationService, error) {
	messageFormat, err := mf.New(locale, &mf.MessageFormatOptions{
		Currency:   "USD",
		ReturnType: mf.ReturnTypeString,
	})
	if err != nil {
		return nil, err
	}

	return &NotificationService{
		messageFormat: messageFormat,
	}, nil
}

// CartMessage generates shopping cart status messages
func (ns *NotificationService) CartMessage(itemCount int, customerGender string) (string, error) {
	template := `{gender, select,
		male {{itemCount, plural, =0 {He has no items in his cart} =1 {He has {itemCount} item in his cart} other {He has {itemCount} items in his cart}}}
		female {{itemCount, plural, =0 {She has no items in her cart} =1 {She has {itemCount} item in her cart} other {She has {itemCount} items in her cart}}}
		other {{itemCount, plural, =0 {They have no items in their cart} =1 {They have {itemCount} item in their cart} other {They have {itemCount} items in their cart}}}
	}`

	msg, err := ns.messageFormat.Compile(template)
	if err != nil {
		return "", err
	}

	result, err := msg(map[string]interface{}{
		"gender":    customerGender,
		"itemCount": itemCount,
	})
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// OrderStatusMessage generates order status notifications
func (ns *NotificationService) OrderStatusMessage(order Order) (string, error) {
	template := `{status, select,
		pending {Order #{orderID} is being processed. {itemCount, plural, one {# item} other {# items}} for a total of ${total}.}
		shipped {Great news {customer}! Your order #{orderID} with {itemCount, plural, one {# item} other {# items}} has been shipped.}
		delivered {Order #{orderID} has been delivered to {customer}. Thank you for your purchase of {itemCount, plural, one {# item} other {# items}}!}
		cancelled {Order #{orderID} has been cancelled. {itemCount, plural, one {# item} other {# items}} will be refunded to your account.}
		other {
			Order #{orderID} status: {status}
		}
	}`

	msg, err := ns.messageFormat.Compile(template)
	if err != nil {
		return "", err
	}

	result, err := msg(map[string]interface{}{
		"status":    order.Status,
		"orderID":   order.ID,
		"customer":  order.Customer,
		"itemCount": order.Items,
		"total":     order.Total,
	})
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// InventoryMessage generates inventory level alerts
func (ns *NotificationService) InventoryMessage(productName string, stockLevel int) (string, error) {
	template := `{stock, plural,
		=0 {⚠️ {product} is out of stock}
		=1 {⚠️ Only {stock} {product} left in stock!}
		other {✅ {stock} {product} items available}
	}`

	msg, err := ns.messageFormat.Compile(template)
	if err != nil {
		return "", err
	}

	result, err := msg(map[string]interface{}{
		"product": productName,
		"stock":   stockLevel,
	})
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

func main() {
	fmt.Println("=== E-commerce MessageFormat Examples ===")

	// Initialize notification service
	notificationService, err := NewNotificationService("en")
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: Shopping Cart Messages
	fmt.Println("\n1. Shopping Cart Messages:")
	cartScenarios := []struct {
		itemCount int
		gender    string
	}{
		{0, "male"},
		{1, "female"},
		{5, "other"},
		{3, "unknown"}, // Will fall back to "other"
	}

	for _, scenario := range cartScenarios {
		message, err := notificationService.CartMessage(scenario.itemCount, scenario.gender)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("   Items: %d, Gender: %s -> %s\n", scenario.itemCount, scenario.gender, message)
	}

	// Example 2: Order Status Messages
	fmt.Println("\n2. Order Status Messages:")
	orders := []Order{
		{"ORD-001", 3, 99.99, "Alice Johnson", "pending"},
		{"ORD-002", 1, 29.99, "Bob Smith", "shipped"},
		{"ORD-003", 5, 149.95, "Charlie Brown", "delivered"},
		{"ORD-004", 2, 59.98, "Dana White", "cancelled"},
	}

	for _, order := range orders {
		message, err := notificationService.OrderStatusMessage(order)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("   %s: %s\n", order.ID, message)
	}

	// Example 3: Inventory Alerts
	fmt.Println("\n3. Inventory Alerts:")
	products := []struct {
		name  string
		stock int
	}{
		{"Premium Headphones", 0},
		{"Wireless Mouse", 1},
		{"USB Cable", 15},
		{"Laptop Stand", 3},
	}

	for _, product := range products {
		message, err := notificationService.InventoryMessage(product.name, product.stock)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("   %s: %s\n", product.name, message)
	}

	fmt.Println("\n=== E-commerce examples completed successfully! ===")
}
