package adapters

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/common/tracing"
	"github.com/sirupsen/logrus"
)

type OrderGrpc struct {
	client orderpb.OrderServiceClient
}

func NewOrderGrpc(client orderpb.OrderServiceClient) *OrderGrpc {
	return &OrderGrpc{client: client}
}

func (o OrderGrpc) UpdateOrder(ctx context.Context, order *orderpb.Order) error {
	ctx, span := tracing.Start(ctx, "order_grpc.update_order")
	defer span.End()

	_, err := o.client.UpdateOrder(ctx, order)
	logrus.Infof("payment_adapter||update_order,err=%v", err)
	return err
}
