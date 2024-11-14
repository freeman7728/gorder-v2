package app

import "github.com/freeman7728/gorder-v2/payment/app/command"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreatePayment command.CreatePaymentHandler
}

type Queries struct {
}
