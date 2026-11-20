package service

import (
	"context"

	"pinshaoshao/internal/database"
)

type Transactor interface {
	WithinTx(ctx context.Context, fn func(context.Context, database.DBTX) error) error
}
