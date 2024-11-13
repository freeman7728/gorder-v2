package main

import (
	"github.com/freeman7728/gorder-v2/common/broker"
	"github.com/freeman7728/gorder-v2/common/config"
	"github.com/freeman7728/gorder-v2/common/logging"
	"github.com/freeman7728/gorder-v2/common/server"
	"github.com/freeman7728/gorder-v2/payment/infrastructure/consumer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.Sub("payment").GetString("service-name")
	serverType := viper.Sub("payment").GetString("server-to-run")

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

	go consumer.NewConsumer().Listen(ch)

	paymentHandler := NewPaymentHandler()
	switch serverType {
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported serviceType")
	default:
		logrus.Panic("unsupported serviceType")
	}

}
