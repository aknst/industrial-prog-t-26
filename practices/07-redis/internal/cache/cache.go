package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrMiss возвращается, если ключа нет в Redis или его TTL истёк.
var ErrMiss = errors.New("cache miss")

const opTimeout = 2 * time.Second

type Cache struct {
	rdb *redis.Client
}

func New(addr, password string, db int) *Cache {
	return &Cache{rdb: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
}

func (c *Cache) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return c.rdb.Ping(ctx).Err()
}

func (c *Cache) Close() error { return c.rdb.Close() }

// Set сохраняет значение; ttl == 0 означает «без срока жизни».
func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	return val, err
}

// TTL возвращает оставшееся время жизни. Для ключа без срока жизни результат
// равен -1, для отсутствующего ключа возвращается ErrMiss.
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl == -2 {
		return 0, ErrMiss
	}
	if ttl == -1 {
		return -1, nil
	}
	return ttl, nil
}

func (c *Cache) Del(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	n, err := c.rdb.Del(ctx, key).Result()
	return n > 0, err
}

// GetOrLoad реализует схему cache-aside: сначала читает кэш, при промахе вызывает
// load (например, запрос к БД), кладёт результат в кэш на ttl и возвращает его.
// Второй результат показывает, что значение было взято из кэша.
func (c *Cache) GetOrLoad(ctx context.Context, key string, ttl time.Duration, load func(context.Context) (string, error)) (string, bool, error) {
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, true, nil
	}
	if !errors.Is(err, ErrMiss) {
		return "", false, err
	}
	val, err = load(ctx)
	if err != nil {
		return "", false, err
	}
	if err := c.Set(ctx, key, val, ttl); err != nil {
		return "", false, err
	}
	return val, false, nil
}
