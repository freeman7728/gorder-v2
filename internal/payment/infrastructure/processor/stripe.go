package processor

import (
	"context"
	"encoding/json"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type StripeProcessor struct {
	apiKey string
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	if apiKey == "" {
		panic("empty api key")
	}
	stripe.Key = apiKey
	return &StripeProcessor{apiKey: apiKey}
}

var successUrl = "http://localhost:8282"

func (s StripeProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {
	// Set your secret key. Remember to switch to your live secret key in production.
	// See your keys here: https://dashboard.stripe.com/apikeys
	//stripe.Key = "sk_test_4eC39HqLyjWDarjtT1zdp7dc"
	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			//Price:    stripe.String(item.PriceID),
			Price:    stripe.String("price_1QLHRJEDLpH1wCU8fOBhRq2m"),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}
	marshalledItems, _ := json.Marshal(items)
	metadata := map[string]string{
		"orderID":    order.ID,
		"customerID": order.CustomerID,
		"status":     order.Status,
		"items":      string(marshalledItems),
	}
	params := &stripe.CheckoutSessionParams{
		Metadata:   metadata,
		LineItems:  items,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successUrl),
	}
	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
