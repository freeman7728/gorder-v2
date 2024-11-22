package tests

import (
	"context"
	"fmt"
	sw "github.com/freeman7728/gorder-v2/common/client/order"
	_ "github.com/freeman7728/gorder-v2/common/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

var (
	ctx    = context.Background()
	server = fmt.Sprintf("http://%s/api", viper.GetString("order.http-addr"))
	client *sw.ClientWithResponses
)

func TestMain(m *testing.M) {
	before()
	m.Run()
}

func before() {
	log.Printf("server=%s\n", server)
	client, _ = sw.NewClientWithResponses(server)
}

func getResponse(t *testing.T, customerId string, body sw.PostCustomerCustomerIdOrdersJSONRequestBody) *sw.PostCustomerCustomerIdOrdersResponse {
	resp, err := client.PostCustomerCustomerIdOrdersWithResponse(ctx, customerId, body)
	if err != nil {
		t.Error(err)
		return nil
	}
	return resp
}

func TestCreateOrder_invalidParams(t *testing.T) {
	response := getResponse(t, "123", sw.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: "123",
		Items:      nil,
	})
	assert.Equal(t, 200, response.StatusCode())
	assert.Equal(t, 2, response.JSON200.Errno)
}

func TestCreateOrder_validParams(t *testing.T) {
	response := getResponse(t, "123", sw.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: "123",
		Items: []sw.ItemWithQuantity{
			{
				Id:       "prod_RCftmpMkL9wnSM",
				Quantity: 10,
			},
		},
	})
	t.Logf("body = %s", string(response.Body))
	assert.Equal(t, 200, response.StatusCode())
	assert.Equal(t, 0, response.JSON200.Errno)
}
