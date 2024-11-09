package adapters

import (
	"context"
	domain "github.com/freeman7728/gorder-v2/order/domain/order"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type MemoryOrderRepository struct {
	lock  sync.RWMutex
	store []*domain.Order
}

var fakeData = []*domain.Order{}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	s := make([]*domain.Order, 0)
	s = append(s, &domain.Order{
		ID:          "fake_id",
		CustomerID:  "fake_customerID",
		Status:      "fake_status",
		PaymentLink: "fake_link",
		Items:       nil,
	})
	return &MemoryOrderRepository{
		lock:  sync.RWMutex{},
		store: s,
	}
}

func (m MemoryOrderRepository) Create(_ context.Context, o *domain.Order) (*domain.Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
	}
	m.store = append(m.store, newOrder)
	logrus.WithFields(logrus.Fields{
		"input_order":        o,
		"store_after_create": m.store,
	}).Debug("memory_order_repo_created")
	return newOrder, nil
}

func (m MemoryOrderRepository) Get(_ context.Context, id, customerID string) (*domain.Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			logrus.Debugf("memory_order_repo_found||found||id=%s||customerID=%s||res=%v", id, customerID, *o)
			return o, nil
		}
	}
	return nil, domain.NotFoundError{OrderID: id}
}

func (m MemoryOrderRepository) Update(ctx context.Context, o *domain.Order, updateFn func(context.Context, *domain.Order) (*domain.Order, error)) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	found := false
	for i, o := range m.store {
		if o.ID == o.ID && o.CustomerID == o.CustomerID {
			found = true
			updateOrder, err := updateFn(ctx, o)
			if err != nil {
				return err
			}
			m.store[i] = updateOrder
		}
	}
	if !found {
		return domain.NotFoundError{OrderID: o.ID}
	}
	return nil
}
