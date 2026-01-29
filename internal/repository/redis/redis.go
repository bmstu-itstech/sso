package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	TTL    time.Duration
}

func New(cfg config.Config) *Redis {
	ctx := context.Background()
	fmt.Println("Connecting to Redis...")
	fmt.Println(cfg)
	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password: "",
		DB:       0,
	})

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Redis errors/ no ping: %v\n", err))
	}
	fmt.Println("Redis ping: ", pong)
	return &Redis{
		client: rdb,
		TTL:    cfg.JWT.TokenTTL,
	}
}

func (r *Redis) userKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

func (r *Redis) HasUserChanges(ctx context.Context, userID int64) (bool, error) {
	key := r.userKey(userID)
	_, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (r *Redis) PublishUserUpdated(ctx context.Context, userID int64) error {
	key := r.userKey(userID)
	err := r.client.Set(ctx, key, 0, r.TTL).Err()
	return err
}
