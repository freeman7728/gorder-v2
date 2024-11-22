package integration

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/product"
)

type StripeAPI struct {
	apiKey string
	stripe *stripe.Price
}

func NewStripeAPI() *StripeAPI {
	key := viper.GetString("stripe-key")
	if key == "" {
		logrus.Fatal("empty_stripe_key")
	}
	return &StripeAPI{
		apiKey: key,
	}
}

func (s *StripeAPI) GetPriceByProductID(ctx context.Context, pid string) (string, error) {
	stripe.Key = s.apiKey
	result, err := product.Get(pid, &stripe.ProductParams{})
	if err != nil {
		return "", err
	}
	return result.DefaultPrice.ID, nil
}
