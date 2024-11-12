package discovery

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/discovery/consul"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

func RegisterToConsul(ctx context.Context, serviceName string) (func() error, error) {
	registry, err := consul.New(viper.GetString("consul.addr"))
	if err != nil {
		return func() error {
			return nil
		}, err
	}
	instanceID := GenerateInstanceID(serviceName)
	grpcAddr := viper.Sub(serviceName).GetString("grpc-addr")
	err = registry.Register(ctx, instanceID, serviceName, grpcAddr)
	if err != nil {
		return func() error {
			return nil
		}, err
	}
	go func() {
		for {
			err := registry.HealthCheck(instanceID, serviceName)
			if err != nil {
				logrus.Panicf("no heartbeat from %s to registry,err=%v", serviceName, err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
	logrus.WithFields(logrus.Fields{
		"serviceName": serviceName,
		"addr":        grpcAddr,
	}).Info("registered to consul")
	return func() error {
		return registry.Deregister(ctx, instanceID, serviceName)
	}, nil
}
