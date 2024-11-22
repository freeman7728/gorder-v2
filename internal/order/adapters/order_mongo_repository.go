package adapters

import (
	"context"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	"github.com/freeman7728/gorder-v2/order/entity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	dBName   = viper.GetString("mongo.db-name")
	collName = viper.GetString("mongo.coll-name")
)

type OrderRepositoryMongo struct {
	db *mongo.Client
}

type orderModel struct {
	MongoID     primitive.ObjectID `bson:"_id"`
	ID          string             `bson:"id"`
	CustomerID  string             `bson:"customer_id"`
	Status      string             `bson:"status"`
	PaymentLink string             `bson:"payment_link"`
	Items       []*entity.Item     `bson:"items"`
}

func NewOrderRepositoryMongo(db *mongo.Client) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{db: db}
}

func (r *OrderRepositoryMongo) collection() *mongo.Collection {
	return r.db.Database(dBName).Collection(collName)
}

func (r *OrderRepositoryMongo) logWithTag(tag string, err error, result any) {
	l := logrus.WithFields(logrus.Fields{
		"tag":            tag,
		"performed_time": time.Now().Unix(),
		"err":            err,
		"result":         result,
	})
	if err != nil {
		l.Infof("order_repository_mongo_%s_fail", tag)
	} else {
		l.Infof("order_repository_mongo_%s_success", tag)
	}
}

func (r *OrderRepositoryMongo) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	defer func() {
		r.logWithTag("create", err, created)
	}()

	write := r.marshalToModel(order)
	res, err := r.collection().InsertOne(ctx, write)
	if err != nil {
		return nil, err
	}
	created = order
	created.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func (r *OrderRepositoryMongo) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	defer func() {
		r.logWithTag("get", err, got)
	}()
	read := &orderModel{}
	mongoID, _ := primitive.ObjectIDFromHex(id)
	cond := bson.M{"_id": mongoID, "customer_id": customerID}
	err = r.collection().FindOne(ctx, cond).Decode(read)
	if err != nil {
		return
	}
	if read == nil {
		return nil, domain.NotFoundError{OrderID: id}
	}
	return r.unmarshal(read), nil
}

// Update 先查找对应的order,然后apply updateFn,再写入回去
func (r *OrderRepositoryMongo) Update(
	ctx context.Context,
	order *domain.Order,
	updateFn func(context.Context, *domain.Order,
	) (*domain.Order, error)) (err error) {
	defer r.logWithTag("after_update", err, nil)
	if order == nil {
		panic("got nil order")
	}
	session, err := r.db.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	err = session.StartTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = session.CommitTransaction(ctx)
		} else {
			_ = session.AbortTransaction(ctx)
		}
	}()
	//inside Transaction
	oldOrder, err := r.Get(ctx, order.ID, order.CustomerID)
	if err != nil {
		return err
	}
	updated, err := updateFn(ctx, oldOrder)
	if err != nil {
		return err
	}
	mongoID, err := primitive.ObjectIDFromHex(oldOrder.ID)
	if err != nil {
		return err
	}
	res, err := r.collection().UpdateOne(
		ctx,
		bson.M{"_id": mongoID, "customer_id": updated.CustomerID},
		bson.M{"$set": bson.M{
			"status":       order.Status,
			"payment_link": order.PaymentLink,
		}},
	)
	if err != nil {
		return err
	}
	r.logWithTag("finish_update", err, res)
	return
}

func (r *OrderRepositoryMongo) marshalToModel(order *domain.Order) *orderModel {
	return &orderModel{
		MongoID:     primitive.NewObjectID(),
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
}

func (r *OrderRepositoryMongo) unmarshal(m *orderModel) *domain.Order {
	return &domain.Order{
		ID:          m.MongoID.Hex(),
		CustomerID:  m.CustomerID,
		Status:      m.Status,
		PaymentLink: m.PaymentLink,
		Items:       m.Items,
	}
}
