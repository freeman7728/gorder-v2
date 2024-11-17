package main

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/config"
	"github.com/freeman7728/gorder-v2/common/discovery"
	"github.com/freeman7728/gorder-v2/common/genproto/stockpb"
	"github.com/freeman7728/gorder-v2/common/logging"
	"github.com/freeman7728/gorder-v2/common/server"
	"github.com/freeman7728/gorder-v2/common/tracing"
	"github.com/freeman7728/gorder-v2/stock/ports"
	"github.com/freeman7728/gorder-v2/stock/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serviceType := viper.GetString("stock.server-to-run")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	defer shutdown(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	//创建application，传给svc
	application := service.NewApplication(ctx)

	deRegister, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = deRegister()
	}()

	switch serviceType {
	case "http":
		//TODO: Temporary disable
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer(application)
			stockpb.RegisterStockServiceServer(server, svc)
		})
	default:
		panic("unexpected service type: " + serviceType)
	}
}
