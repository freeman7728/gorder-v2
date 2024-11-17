package query

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/decorator"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type GetCustomerOrder struct {
	CustomerID string
	OrderID    string
}

type GetCustomerOrderHandler decorator.QueryHandler[GetCustomerOrder, *domain.Order]

type getCustomerOrderHandler struct {
	orderRepo domain.Repository
}

func NewGetCustomerOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) GetCustomerOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	return decorator.ApplyQueryDecorators[GetCustomerOrder, *domain.Order](
		getCustomerOrderHandler{orderRepo: orderRepo},
		logger,
		metricsClient,
	)
}

func (g getCustomerOrderHandler) Handle(ctx context.Context, query GetCustomerOrder) (*domain.Order, error) {
	t := otel.Tracer("getCustomerOrderHandler")
	ctx, span := t.Start(ctx, "getCustomerOrderHandler")
	defer span.End()
	o, err := g.orderRepo.Get(ctx, query.OrderID, query.CustomerID)
	if err != nil {
		return nil, err
	}
	return o, nil
}
