package service

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/client"
	"github.com/freeman7728/gorder-v2/common/metrics"
	"github.com/freeman7728/gorder-v2/payment/adapters"
	"github.com/freeman7728/gorder-v2/payment/app"
	"github.com/freeman7728/gorder-v2/payment/app/command"
	"github.com/freeman7728/gorder-v2/payment/domain"
	"github.com/freeman7728/gorder-v2/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	orderGRPCClient, closeOrderClient, err := client.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	orderGrpc := adapters.NewOrderGrpc(orderGRPCClient)
	//inmemProcessor := processor.NewInmemProcessor()
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("stripe-key"))
	return newApplication(ctx, orderGrpc, stripeProcessor), func() { _ = closeOrderClient() }
}
func newApplication(ctx context.Context, grpc command.OrderService, inmemProcessor domain.Processor) app.Application {
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(inmemProcessor, grpc, metricsClient, logger),
		},
		Queries: app.Queries{},
	}
}
