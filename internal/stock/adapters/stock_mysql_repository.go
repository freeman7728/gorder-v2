package adapters

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/stock/infrastructure/persistent"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

func (m RepositoryStockMysql) UpdateStock(
	ctx context.Context,
	data []*orderpb.ItemWithQuantity,
	updateFn func(
		ctx context.Context,
		existing []*orderpb.ItemWithQuantity,
		query []*orderpb.ItemWithQuantity,
	) ([]*orderpb.ItemWithQuantity, error),
) error {
	return m.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				logrus.Warnf("update stock transaction err=%v", err)
			}
		}()
		var dest []*persistent.StockModel
		if err = tx.Table("o_stock").Where("product_id IN ?", getIDFromEntities(data)).Find(&dest).Error; err != nil {
			return err
		}
		existing := m.unmarshalFromDatabase(dest)

		updated, err := updateFn(ctx, existing, data)
		if err != nil {
			return err
		}

		for _, upd := range updated {
			if err = tx.Table("o_stock").Where("product_id = ?", upd.ID).Update("quantity", upd.Quantity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (m RepositoryStockMysql) unmarshalFromDatabase(dest []*persistent.StockModel) []*orderpb.ItemWithQuantity {
	var result []*orderpb.ItemWithQuantity
	for _, i := range dest {
		result = append(result, &orderpb.ItemWithQuantity{
			ID:       i.ProductID,
			Quantity: i.Quantity,
		})
	}
	return result
}

func getIDFromEntities(items []*orderpb.ItemWithQuantity) []string {
	var ids []string
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	return ids
}
