package service

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/metrics"
	"github.com/freeman7728/gorder-v2/stock/adapters"
	"github.com/freeman7728/gorder-v2/stock/app"
	"github.com/freeman7728/gorder-v2/stock/app/query"
	"github.com/freeman7728/gorder-v2/stock/infrastructure/integration"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	stockRepo := adapters.NewMemoryStockRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	stripeAPI := integration.NewStripeAPI()
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, logger, metricsClient, stripeAPI),
		},
	}
}
