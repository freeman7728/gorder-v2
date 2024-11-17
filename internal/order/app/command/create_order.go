package command

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/freeman7728/gorder-v2/common/broker"
	"github.com/freeman7728/gorder-v2/common/decorator"
	"github.com/freeman7728/gorder-v2/order/app/query"
	"github.com/freeman7728/gorder-v2/order/convertor"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	"github.com/freeman7728/gorder-v2/order/entity"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*entity.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

// 取别名是为了更简洁直观地定义返回值
type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	stockGRPC query.StockService
	channel   *amqp.Channel
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
	stockGRPC query.StockService,
	channel *amqp.Channel,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	if logger == nil {
		panic("nil logger")
	}
	if stockGRPC == nil {
		panic("nil stockGRPC")
	}
	if channel == nil {
		panic("nil channel")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{
			orderRepo: orderRepo,
			stockGRPC: stockGRPC,
			channel:   channel,
		},
		logger,
		metricsClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
	}
	o, err := c.orderRepo.Create(ctx,
		&domain.Order{
			CustomerID: cmd.CustomerID,
			Items:      validItems,
		})
	logrus.Infof("input_order=%v", o)
	if err != nil {
		return nil, err
	}

	//订单成功创建之后，绑定queue到exchange
	q, err := c.channel.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	marshaledOrder, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	err = c.channel.PublishWithContext(
		ctx, "", q.Name, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         marshaledOrder,
		},
	)
	if err != nil {
		return nil, err
	}
	return &CreateOrderResult{OrderID: o.ID}, nil
}

func (c createOrderHandler) validate(ctx context.Context, items []*entity.ItemWithQuantity) ([]*entity.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have one item")
	}
	//去重，也就是打包
	items = packItems(items)
	//检查仓库中是否有物品的数量小于订单的要求
	resp, err := c.stockGRPC.CheckIfItemInStock(ctx, convertor.NewItemWithQuantityConvertor().EntitiesToProtos(items))
	if err != nil {
		return nil, err
	}
	return convertor.NewItemConvertor().ProtosToEntities(resp.Items), nil
}

func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}
	items = make([]*entity.ItemWithQuantity, 0)
	for ID, quantity := range merged {
		items = append(items, &entity.ItemWithQuantity{ID: ID, Quantity: quantity})
	}
	return items
}
