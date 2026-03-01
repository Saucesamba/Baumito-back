package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type AdCacheRepository struct {
	client *redis.Client
}

func NewAdCacheRepository(addr string) *AdCacheRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &AdCacheRepository{client: rdb}
}

// IncrementViews увеличивает счетчик в памяти (очень быстро)
func (r *AdCacheRepository) IncrementViews(ctx context.Context, adID string) error {
	key := fmt.Sprintf("ad:views:%s", adID)
	return r.client.Incr(ctx, key).Err()
}

// GetViews возвращает текущее количество просмотров из кэша
func (r *AdCacheRepository) GetViews(ctx context.Context, adID string) (int64, error) {
	key := fmt.Sprintf("ad:views:%s", adID)
	return r.client.Get(ctx, key).Int64()
}
