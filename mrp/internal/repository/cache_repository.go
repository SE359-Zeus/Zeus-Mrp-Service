package repository

import (
	"context"
)

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string, dest interface{}) error
}

// In the service, you would inject this:
// type productionService struct {
//    repo  repository.MRPRepository
//    cache repository.CacheRepository
// }
