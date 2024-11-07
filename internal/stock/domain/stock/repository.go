package stock

import (
	"context"
	"fmt"
	"github.com/freeman7728/gorder-v2/common/genproto/orderpb"
	"strings"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error)
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("These Items with ID %s not found in stock", strings.Join(e.Missing, ","))
}
