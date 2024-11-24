package query

import (
	"context"
	_ "github.com/freeman7728/gorder-v2/common/config"
	"github.com/freeman7728/gorder-v2/common/decorator"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/common/handler/redis"
	domain "github.com/freeman7728/gorder-v2/stock/domain/stock"
	"github.com/freeman7728/gorder-v2/stock/infrastructure/integration"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	RedisLockPrefix = "check_stock"
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
	if err := lock(ctx, getLockKey(query)); err != nil {
		return nil, err
	}
	defer func() {
		if err := unlock(ctx, getLockKey(query)); err != nil {
			logrus.Warnf("unlock failed: %v", err)
		}
	}()
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
	//扣库存
	if err := h.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	return res, nil
}

func getLockKey(query CheckIfItemsInStock) string {
	ids := make([]string, 0)
	for _, item := range query.Items {
		ids = append(ids, item.ID)
	}
	return RedisLockPrefix + strings.Join(ids, "_")
}

func unlock(ctx context.Context, key string) error {
	return redis.Del(ctx, redis.LocalClient(), key)
}

func lock(ctx context.Context, key string) error {
	return redis.SetNX(ctx, redis.LocalClient(), key, "1", 5*time.Minute)
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
		return h.stockRepo.UpdateStock(ctx, items, func(
			ctx context.Context,
			existing []*orderpb.ItemWithQuantity,
			query []*orderpb.ItemWithQuantity,
		) ([]*orderpb.ItemWithQuantity, error) {
			var newItems []*orderpb.ItemWithQuantity
			for _, e := range existing {
				for _, q := range query {
					if e.ID == q.ID {
						newItems = append(newItems, &orderpb.ItemWithQuantity{
							ID:       e.ID,
							Quantity: e.Quantity - q.Quantity,
						})
					}
				}
			}
			return newItems, nil
		})
	}
	return domain.ExceedStockError{FailedOn: exceedDetail}
}
