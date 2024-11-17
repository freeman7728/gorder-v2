package ports

import (
	context "context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"github.com/freeman7728/gorder-v2/order/app"
	"github.com/freeman7728/gorder-v2/order/app/command"
	"github.com/freeman7728/gorder-v2/order/app/query"
	"github.com/freeman7728/gorder-v2/order/convertor"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) CreateOrder(ctx context.Context, request *orderpb.CreateOrderRequest) (*emptypb.Empty, error) {
	_, err := G.app.Commands.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: request.CustomerID,
		Items:      convertor.NewItemWithQuantityConvertor().ProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (G GRPCServer) GetOrder(ctx context.Context, request *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	//调用repo
	o, err := G.app.Queries.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: request.CustomerID,
		OrderID:    request.OrderID,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return convertor.NewOrderConvertor().EntityToProto(o), err
}

func (G GRPCServer) UpdateOrder(ctx context.Context, request *orderpb.Order) (_ *emptypb.Empty, err error) {
	t := otel.Tracer("UpdateOrder")
	ctx, span := t.Start(ctx, "UpdateOrder")
	defer span.End()
	logrus.Infof("order_grpc||request_in||request=%v", request)
	order, err := domain.NewOrder(
		request.ID,
		request.CustomerID,
		request.Status,
		request.PaymentLink,
		convertor.NewItemConvertor().ProtosToEntities(request.Items))
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}
	_, err = G.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: order,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			return order, nil
		},
	})
	return
}
