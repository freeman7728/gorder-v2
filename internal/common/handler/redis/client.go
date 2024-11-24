package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

func SetNX(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start": now,
			"key":   key,
			"value": value,
			"error": err,
			"cost":  time.Since(now).Milliseconds(),
		})
		if err == nil {
			l.Info("redis setnx success")
		} else {
			l.Info("redis setnx error")
		}
	}()
	if client == nil {
		return errors.New("redis_client is nil")
	}
	_, err = client.SetNX(ctx, key, value, ttl).Result()
	return err
}
func Del(ctx context.Context, client *redis.Client, key string) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start": now,
			"key":   key,
			"error": err,
			"cost":  time.Since(now).Milliseconds(),
		})
		if err == nil {
			l.Info("redis del success")
		} else {
			l.Info("redis del error")
		}
	}()
	if client == nil {
		return errors.New("redis_client is nil")
	}
	_, err = client.Del(ctx, key).Result()
	return err
}
