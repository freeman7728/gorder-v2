package command

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/decorator"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*orderpb.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	//stockGRPC
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{orderRepo: orderRepo},
		logger,
		metricsClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	//TODO: Call stockGRPC to get Items
	var stockResponse []*orderpb.Item
	for _, item := range cmd.Items {
		stockResponse = append(stockResponse, &orderpb.Item{
			ID:       item.ID,
			Quantity: item.Quantity,
		})
	}
	o, err := c.orderRepo.Create(ctx,
		&domain.Order{
			CustomerID: cmd.CustomerID,
			Items:      stockResponse,
		})
	if err != nil {
		return nil, err
	}
	return &CreateOrderResult{OrderID: o.ID}, nil
}
