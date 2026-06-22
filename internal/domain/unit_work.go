package domain

import (
	"context"
)

type UnitOfWork interface {
	Tx(ctx context.Context, atomicFn func(ctx context.Context) error) error
}
