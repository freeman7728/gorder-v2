package stock

import (
	"context"
	"fmt"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"strings"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error)
	GetStocks(ctx context.Context, ids []string) ([]*orderpb.ItemWithQuantity, error)
	//CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) ([]*orderpb.Item, error)
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("These Items with ID %s not found in stock", strings.Join(e.Missing, ","))
}

type ExceptionalItem struct {
	Id   string
	Want int32
	Have int32
}

type ExceedStockError struct {
	FailedOn []ExceptionalItem
}

func (e ExceedStockError) Error() string {
	var res []string
	for _, item := range e.FailedOn {
		res = append(res, fmt.Sprintf("product %s want %d but have %d", item.Id, item.Want, item.Have))
	}
	return fmt.Sprintf("Product ID %s exceeds Stock", strings.Join(res, ","))
}
