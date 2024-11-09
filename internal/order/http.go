package main

import (
	"github.com/freeman7728/gorder-v2/order/app"
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
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) GetCustomerCustomerIDOrderOrderID(c *gin.Context, customerID string, orderID string) {
	o, err := H.app.Queries.GetCustomerOrder.Handle(c, query.GetCustomerOrder{
		CustomerID: "fake_customerID",
		OrderID:    "fake_id",
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": o})
}
