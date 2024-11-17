package main

import (
	"fmt"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/common/tracing"
	"github.com/freeman7728/gorder-v2/order/app"
	"github.com/freeman7728/gorder-v2/order/app/command"
	"github.com/freeman7728/gorder-v2/order/app/query"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HTTPServer struct {
	app app.Application
}

func NewHTTPServer(app app.Application) *HTTPServer {
	return &HTTPServer{app: app}
}

func (H HTTPServer) PostCustomerCustomerIDOrder(c *gin.Context, customerID string) {
	ctx, span := tracing.Start(c, "PostCustomerCustomerIDOrder")
	defer span.End()
	var req orderpb.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	result, err := H.app.Commands.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: req.CustomerID,
		Items:      req.Items,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
		return
	}
	traceID := tracing.TraceID(ctx)
	c.JSON(http.StatusOK, gin.H{
		"message":      "success",
		"trace_id":     traceID,
		"customer_id":  req.CustomerID,
		"order_id":     result.OrderID,
		"redirect_url": fmt.Sprintf("http://centos2:8282/success?customerID=%s&orderID=%s", req.CustomerID, result.OrderID),
	})
}

func (H HTTPServer) GetCustomerCustomerIDOrderOrderID(c *gin.Context, customerID string, orderID string) {
	ctx, span := tracing.Start(c, "GetCustomerCustomerIDOrderOrderID")
	defer span.End()
	o, err := H.app.Queries.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
		return
	}
	traceID := tracing.TraceID(ctx)
	c.JSON(http.StatusOK, gin.H{
		"message":  "success",
		"trace_id": traceID,
		"data": gin.H{
			"Order": o,
		},
	})
}
