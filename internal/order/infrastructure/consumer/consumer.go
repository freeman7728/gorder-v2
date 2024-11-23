package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/freeman7728/gorder-v2/common/broker"
	"github.com/freeman7728/gorder-v2/order/app"
	"github.com/freeman7728/gorder-v2/order/app/command"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{app: app}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderPaid, true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	forever := make(chan bool)
	go func() {
		for {
			for msg := range msgs {
				c.handleMessage(msg, q, ch)
			}
		}
	}()
	<-forever
}

// 接收到成功消费事件之后，去更新订单的状态
func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue, channel *amqp.Channel) {
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	_, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()

	o := &domain.Order{}
	err = json.Unmarshal(msg.Body, o)
	if err != nil {
		logrus.Infof("error unmarshalling msg to domain.Order: %s", err)
		_ = msg.Nack(false, false)
		return
	}
	_, err = c.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err := o.IsPaid(); err != nil {
				return nil, err
			}
			return order, nil
		},
	})
	if err != nil {
		logrus.Infof("error updating order,orderID=%s,err= %v", o.ID, err)
		err := broker.HandleRetry(ctx, channel, &msg)
		if err != nil {
			logrus.Warnf("retry_error,error Handling retry ,messageID=%s,err=%v", msg.MessageId, err)
		}
		return
	}
	span.AddEvent("order.updated")
	logrus.Info("order consume paid event success!")
}
