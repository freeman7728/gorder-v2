package main

import (
	"fmt"
	"github.com/freeman7728/gorder-v2/common"
	client "github.com/freeman7728/gorder-v2/common/client/order"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/order/app"
	"github.com/freeman7728/gorder-v2/order/app/command"
	"github.com/freeman7728/gorder-v2/order/app/query"
	"github.com/freeman7728/gorder-v2/order/convertor"
	_ "github.com/freeman7728/gorder-v2/order/convertor"
	"github.com/gin-gonic/gin"
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
		req  orderpb.CreateOrderRequest
		err  error
		resp struct {
			CustomerID  string `json:"customer_id"`
			OrderID     string `json:"order_id"`
			RedirectURL string `json:"redirect_url"`
		}
	)
	defer func() {
		H.Response(c, err, resp)
	}()
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerID,
		Items:      convertor.NewItemWithQuantityConvertor().ProtosToEntities(req.Items),
	})
	if err != nil {
		return
	}
	resp.CustomerID = req.CustomerID
	resp.OrderID = r.OrderID
	resp.RedirectURL = fmt.Sprintf("http://centos2:8282/success?customerID=%s&orderID=%s", req.CustomerID, r.OrderID)
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
