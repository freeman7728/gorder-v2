package main

import (
	"errors"
	"fmt"
	"github.com/freeman7728/gorder-v2/common"
	client "github.com/freeman7728/gorder-v2/common/client/order"
	"github.com/freeman7728/gorder-v2/order/app"
	"github.com/freeman7728/gorder-v2/order/app/command"
	"github.com/freeman7728/gorder-v2/order/app/dto"
	"github.com/freeman7728/gorder-v2/order/app/query"
	"github.com/freeman7728/gorder-v2/order/convertor"
	_ "github.com/freeman7728/gorder-v2/order/convertor"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	app app.Application
	common.BaseResponse
}

func NewHTTPServer(app app.Application) *HTTPServer {
	return &HTTPServer{app: app}
}

func (H HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	var (
		req  client.CreateOrderRequest
		err  error
		resp dto.CreateOrderResponse
	)
	defer func() {
		H.Response(c, err, resp)
	}()
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	if err = H.validate(req); err != nil {
		logrus.Warnf("validate request error: %v", err)
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerId,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items),
	})
	if err != nil {
		return
	}
	resp = dto.CreateOrderResponse{
		OrderID:     r.OrderID,
		CustomerID:  req.CustomerId,
		RedirectURL: fmt.Sprintf("http://centos2:8282/success?customerID=%s&orderID=%s", req.CustomerId, r.OrderID),
	}
}

func (H HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerId string, orderId string) {
	var (
		err  error
		resp struct {
			Order *client.Order
		}
	)
	defer func() {
		H.Response(c, err, resp)
	}()
	o, err := H.app.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerId,
		OrderID:    orderId,
	})
	if err != nil {
		return
	}
	resp.Order = convertor.NewOrderConvertor().EntityToClient(o)
}

func (H HTTPServer) validate(req client.CreateOrderRequest) error {
	for _, v := range req.Items {
		if v.Quantity <= 0 {
			return errors.New("quantity must be positive")
		}
	}
	return nil
}
