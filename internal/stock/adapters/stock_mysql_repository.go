package adapters

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/stock/infrastructure/persistent"
)

type RepositoryStockMysql struct {
	db *persistent.Mysql
}

func NewRepositoryStockMysql(db *persistent.Mysql) *RepositoryStockMysql {
	return &RepositoryStockMysql{db: db}
}

func (m RepositoryStockMysql) GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (m RepositoryStockMysql) GetStocks(ctx context.Context, ids []string) ([]*orderpb.ItemWithQuantity, error) {
	data, err := m.db.BatchGetStockByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	var result []*orderpb.ItemWithQuantity
	for _, d := range data {
		result = append(result, &orderpb.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return result, nil
}
