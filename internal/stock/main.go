package main

import (
	"github.com/freeman7728/gorder-v2/common/config"
	"github.com/freeman7728/gorder-v2/common/genproto/stockpb"
	"github.com/freeman7728/gorder-v2/common/server"
	"github.com/freeman7728/gorder-v2/stock/ports"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serviceType := viper.GetString("stock.server-to-run")
	switch serviceType {
	case "http":
		//TODO: Temporary disable
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer()
			stockpb.RegisterStockServiceServer(server, svc)
		})
	default:
		panic("unexpected service type: " + serviceType)
	}
}
