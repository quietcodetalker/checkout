package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInternal = errors.New("internal error")
	ErrNotFound = errors.New("not found")
)

type Cache interface {
	Set(ctx context.Context, k string, v any, d time.Duration) error
	Get(ctx context.Context, k string) (any, error)
}
