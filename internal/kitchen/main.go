package main

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/broker"
	grpcClient "github.com/freeman7728/gorder-v2/common/client"
	_ "github.com/freeman7728/gorder-v2/common/config"
	"github.com/freeman7728/gorder-v2/common/tracing"
	"github.com/freeman7728/gorder-v2/kitchen/adapters"
	"github.com/freeman7728/gorder-v2/kitchen/infrastructure/consumer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
}

func main() {
	serviceName := viper.GetString("kitchen.service-name")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	defer shutdown(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	orderClient, closeFunc, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer closeFunc()

	orderGRPC := adapters.NewOrderGRPC(orderClient)
	ch, closeCh := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = ch.Close()
		_ = closeCh()
	}()
	forever := make(chan bool)
	go prometheusFunc()
	go consumer.NewConsumer(orderGRPC).Listen(ch)
	<-forever
}
