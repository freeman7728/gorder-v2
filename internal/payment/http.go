package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/freeman7728/gorder-v2/common/broker"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/payment/domain"
	"github.com/freeman7728/gorder-v2/payment/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"
)

type PaymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(channel *amqp.Channel) *PaymentHandler {
	return &PaymentHandler{channel: channel}
}

func (h *PaymentHandler) RegisterRoutes(c *gin.Engine) {
	c.POST("/api/webhook", h.handleWebhook)
	handler := promhttp.HandlerFor(metrics.Reg, promhttp.HandlerOpts{})
	c.GET("/metrics", gin.WrapH(handler))
}

func (h *PaymentHandler) handleWebhook(c *gin.Context) {
	logrus.Infof("webhook called by stripe")
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Infof("Error reading request body: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, err)
		return
	}
	endpointSecret := viper.GetString("ENDPOINT_STRIPE_SECRET")
	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		endpointSecret)

	if err != nil {
		logrus.Infof("Error verifying webhook signature: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err = json.Unmarshal(event.Data.Raw, &session); err != nil {
			logrus.Infof("error unmarshal event.data.raw into session: %v\n", err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			logrus.Infof("payment for checkout session %v success", session.ID)
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			var items []*orderpb.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)
			marshalledOrder, err := json.Marshal(domain.Order{
				ID:          session.Metadata["orderID"],
				CustomerID:  session.Metadata["customerID"],
				Status:      string(stripe.CheckoutSessionPaymentStatusPaid),
				PaymentLink: session.Metadata["paymentLink"],
				Items:       items,
			})
			if err != nil {
				logrus.Infof("Error marshalling order into JSON: %v\n", err)
				c.JSON(http.StatusBadRequest, err.Error())
				return
			}
			tr := otel.Tracer("rabbitmq")
			mqCtx, span := tr.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid))
			defer span.End()

			headers := broker.InjectRabbitMQHeaders(mqCtx)
			err = h.channel.PublishWithContext(ctx, broker.EventOrderPaid, "", false, false, amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         marshalledOrder,
				Headers:      headers,
			})
			if err != nil {
				logrus.Infof("Error publishing order %v: %v\n", session.ID, err)
				c.JSON(http.StatusBadRequest, err.Error())
				return
			}
			logrus.Infof("Successfully published order to%v,body:%v", broker.EventOrderPaid, string(marshalledOrder))
			c.JSON(http.StatusOK, nil)
		}
	default:
		logrus.Infof("Unhandled event type: %s\n", event.Type)
	}

	c.Writer.WriteHeader(http.StatusOK)
}
