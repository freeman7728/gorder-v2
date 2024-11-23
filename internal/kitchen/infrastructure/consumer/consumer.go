package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/freeman7728/gorder-v2/common/broker"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"time"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, request *orderpb.Order) error
}

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
}

type Consumer struct {
	orderGRPC OrderService
}

func NewConsumer(orderGRPC OrderService) *Consumer {
	return &Consumer{orderGRPC: orderGRPC}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s,err=%v", q.Name, err)
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

func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue, channel *amqp.Channel) {
	var err error
	logrus.Infof("Kitchen receive a message from %s,msg=%v", q.Name, string(msg.Body))

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)

	t := otel.Tracer("kitchen_consume_a_msg")
	ctx, span := t.Start(ctx, "kitchen_consume_a_msg")
	defer func() {
		span.End()
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()

	o := &Order{}
	err = json.Unmarshal(msg.Body, o)
	if err != nil {
		logrus.Warnf("fail to unmarshal msg to order,err=%v", err)
		return
	}
	if o.Status != "paid" {
		err = errors.New("order not paid,cannot cook")
	}
	cook(o)
	if err != nil {
		logrus.Warnf("fail to create paymentLink for order,err=%v", err)
		logrus.Infof("error updating order,orderID=%s,err= %v", o.ID, err)
		err := broker.HandleRetry(ctx, channel, &msg)
		if err != nil {
			logrus.Warnf("retry_error,error Handling retry ,messageID=%s,err=%v", msg.MessageId, err)
		}
		return
	}
	span.AddEvent("payment.created")
	logrus.Info("consume order successfully")

	err = c.orderGRPC.UpdateOrder(ctx, &orderpb.Order{
		CustomerID:  o.CustomerID,
		Status:      "ready",
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
		ID:          o.ID,
	})
	if err != nil {
		if err = broker.HandleRetry(ctx, channel, &msg); err != nil {
			logrus.Warnf("kitchen:error handling retry,error=%v", err)
		}
		return
	}
	span.AddEvent("kitchen.order.finish.updated")
	logrus.Infof("consume success")
}

func cook(o *Order) {
	logrus.Infof("chief is cooking %v", o)
	time.Sleep(5 * time.Second)
}
