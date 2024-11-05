package main

import (
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{}
}

func (H HTTPServer) PostCustomerCustomerIDOrder(c *gin.Context, customerID string) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) GetCustomerCustomerIDOrderOrderID(c *gin.Context, customerID string, orderID string) {
	//TODO implement me
	panic("implement me")
}
