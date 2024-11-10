package query

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
)

type StockService interface {
	CheckIfItemInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) error
	GetItems(ctx context.Context, itemsID []string) ([]*orderpb.Item, error)
}
