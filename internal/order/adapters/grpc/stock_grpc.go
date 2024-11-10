package grpc

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/common/genproto/stockpb"
	"github.com/sirupsen/logrus"
)

type StockGRPC struct {
	client stockpb.StockServiceClient
}

func NewStockGRPC(client stockpb.StockServiceClient) *StockGRPC {
	return &StockGRPC{client: client}

}

func (s StockGRPC) CheckIfItemInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) error {
	resp, err := s.client.CheckIfItemsInStock(ctx, &stockpb.CheckIfItemsInStockRequest{Items: items})
	logrus.Info("stock_grpc response", resp)
	return err
}

func (s StockGRPC) GetItems(ctx context.Context, itemsID []string) ([]*orderpb.Item, error) {
	resp, err := s.client.GetItems(ctx, &stockpb.GetItemsRequest{ItemIDs: itemsID})
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}
