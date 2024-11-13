package order

import (
	"errors"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
}

func NewOrder(ID string, customerID string, status string, paymentLink string, items []*orderpb.Item) (*Order, error) {
	//业务逻辑写在domain里面
	if ID == "" {
		return nil, errors.New("empty ID")
	}
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}
	if status == "" {
		return nil, errors.New("empty status")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{
		ID:          ID,
		CustomerID:  customerID,
		Status:      status,
		PaymentLink: "",
		Items:       items,
	}, nil
}

func (o Order) DomainToOrderpb() *orderpb.Order {
	return &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
	}
}

func OrderpbToDomain(order orderpb.Order) *Order {
	return &Order{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
}
