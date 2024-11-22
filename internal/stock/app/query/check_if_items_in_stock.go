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
	return res, nil
}
