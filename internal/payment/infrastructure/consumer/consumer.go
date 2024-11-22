package consumer

import (
	"context"
	"encoding/json"
	"github.com/freeman7728/gorder-v2/common/broker"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/payment/app"
	"github.com/freeman7728/gorder-v2/payment/app/command"
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
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
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
	logrus.Infof("Payment receive a message from %s,msg=%v", q.Name, string(msg.Body))

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)

	t := otel.Tracer("consume")
	ctx, span := t.Start(ctx, "consume")
	defer span.End()

	o := &orderpb.Order{}
	err := json.Unmarshal(msg.Body, o)
	if err != nil {
		logrus.Warnf("fail to unmarshal msg to order,err=%v", err)
		return
	}
	_, err = c.app.Commands.CreatePayment.Handle(ctx, command.CreatePayment{Order: o})
	if err != nil {
		logrus.Warnf("fail to create paymentLink for order,err=%v", err)
		logrus.Infof("error updating order,orderID=%s,err= %v", o.ID, err)
		err := broker.HandleRetry(ctx, channel, &msg)
		if err != nil {
			logrus.Warnf("retry_error,error Handling retry ,messageID=%s,err=%v", msg.MessageId, err)
		}
		_ = msg.Nack(false, false)
		return
	}
	span.AddEvent("payment.created")
	msg.Ack(false)
	logrus.Info("consume order successfully")
}
