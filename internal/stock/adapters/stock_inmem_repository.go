package adapters

import (
	"context"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	domain "github.com/freeman7728/gorder-v2/stock/domain/stock"
	"sync"
)

type MemoryStockRepository struct {
	lock  sync.RWMutex
	store map[string]*orderpb.Item
}

func NewMemoryStockRepository() *MemoryStockRepository {
	return &MemoryStockRepository{
		lock:  sync.RWMutex{},
		store: make(map[string]*orderpb.Item),
	}
}

func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var (
		res     []*orderpb.Item
		missing []string
	)
	for _, id := range ids {
		if item, exist := m.store[id]; exist {
			res = append(res, item)
		} else {
			missing = append(missing, id)
		}
	}
	if len(res) == len(ids) {
		return res, nil
	}
	return res, domain.NotFoundError{Missing: missing}
}
