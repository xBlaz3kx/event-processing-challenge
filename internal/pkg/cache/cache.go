package cache

import "golang.org/x/net/context"

type Cache interface {
	Get(ctx context.Context, key string) (*float64, error)
	Set(ctx context.Context, key string, value float64) error
}
