package query

import (
	"context"
	_ "github.com/freeman7728/gorder-v2/common/config"
	"github.com/freeman7728/gorder-v2/common/decorator"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	domain "github.com/freeman7728/gorder-v2/stock/domain/stock"
	"github.com/freeman7728/gorder-v2/stock/infrastructure/integration"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
	stripeAPI *integration.StripeAPI
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
	stripeAPI *integration.StripeAPI,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	if stripeAPI == nil {
		panic("nil stripeAPI")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStockHandler{stockRepo: stockRepo, stripeAPI: stripeAPI},
		logger,
		metricClient,
	)
}

var stub = map[string]string{
	"1": "price_1QLHRJEDLpH1wCU8fOBhRq2m",
	"2": "price_1QLHBJEDLpH1wCU8GgjhtjRg",
}

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	if err := h.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	var res []*orderpb.Item
	for _, i := range query.Items {
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, i.ID)
		if err != nil {
			logrus.Warnf("GetPriceByProductID ItemID: %s , error: %v,", err, i.ID)
			continue
		}
		res = append(res, &orderpb.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
			PriceID:  priceID,
		})
	}
	//TODO 扣库存
	return res, nil
}

func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, items []*orderpb.ItemWithQuantity) error {
	var ids []string
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	records, err := h.stockRepo.GetStocks(ctx, ids)
	if err != nil {
		return err
	}
	idQuantityMap := make(map[string]int32)
	for _, record := range records {
		idQuantityMap[record.ID] += record.Quantity
	}
	ok := true
	exceedDetail := make([]domain.ExceptionalItem, 0)
	for _, item := range items {
		if idQuantityMap[item.ID] < item.Quantity {
			ok = false
			exceedDetail = append(exceedDetail, domain.ExceptionalItem{
				Id:   item.ID,
				Want: item.Quantity,
				Have: idQuantityMap[item.ID],
			})
		}
	}
	if ok {
		return nil
	}
	return domain.ExceedStockError{FailedOn: exceedDetail}
}
