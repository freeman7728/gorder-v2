package client

import (
	"context"
	"errors"
	"github.com/freeman7728/gorder-v2/common/discovery"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/common/genproto/stockpb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
)

func NewStockGRPCClient(ctx context.Context) (client stockpb.StockServiceClient, close func() error, err error) {
	if !WaitForStockGRPCCLient(viper.GetDuration("dial-grpc-timeout") * time.Second) {
		return nil, nil, errors.New("stock grpc timeout")
	}
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("stock.service-name"))
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for stock grpc")
	}
	opts := grpcDialOpts()
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	return stockpb.NewStockServiceClient(conn), conn.Close, nil
}

func NewOrderGRPCClient(ctx context.Context) (client orderpb.OrderServiceClient, close func() error, err error) {
	if !WaitForOrderGRPCCLient(viper.GetDuration("dial-grpc-timeout") * time.Second) {
		return nil, nil, errors.New("order grpc timeout")
	}
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("order.service-name"))
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	if grpcAddr == "" {
		logrus.Warn("empty grpc addr for order grpc")
	}
	opts := grpcDialOpts()
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error {
			return nil
		}, err
	}
	return orderpb.NewOrderServiceClient(conn), conn.Close, nil
}

func grpcDialOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

func WaitForOrderGRPCCLient(timeout time.Duration) bool {
	logrus.Infof("waiting for order grpc client to connect at %v", timeout)
	return waitFor(viper.GetString("order.grpc-addr"), timeout)
}

func WaitForStockGRPCCLient(timeout time.Duration) bool {
	logrus.Infof("waiting for stock grpc client to connect at %v", timeout)
	return waitFor(viper.GetString("stock.grpc-addr"), timeout)
}

func waitFor(addr string, timeout time.Duration) bool {
	portAvailable := make(chan struct{})
	timeoutCh := time.After(timeout)

	go func() {
		for {
			select {
			case <-timeoutCh:
				return
			default:
			}
			_, err := net.DialTimeout("tcp", addr, time.Second)
			if err == nil {
				close(portAvailable)
				return
			}
			time.Sleep(time.Millisecond * 200)
		}
	}()

	select {
	case <-portAvailable:
		return true
	case <-timeoutCh:
		return false
	}
}
