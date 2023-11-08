package PipBot

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
)

type RedisClient struct {
	*redis.Client
	delimiter string
}

func NewRedisClient(delimiter string) *RedisClient {
	return &RedisClient{
		Client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		delimiter: delimiter,
	}
}

const BatchSize = 100

func (c *RedisClient) Wildcard(name string) string {
	return fmt.Sprintf("%s%s*", name, c.delimiter)
}

func (c *RedisClient) Key(parts ...string) string {
	ret := ""
	for i, part := range parts {
		if i > 0 {
			ret += c.delimiter
		}
		ret += part
	}
	return ret
}

func (c *RedisClient) AllKeys(ctx context.Context, name string) chan string {
	cursor := uint64(0)
	ret := make(chan string)
	go func() {
		defer close(ret)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var keys []string
				keys, cursor = c.Scan(ctx, cursor, c.Wildcard(name), BatchSize).Val()
				for _, key := range keys {
					ret <- key
				}
				if cursor == 0 {
					return
				}
			}
		}
	}()
	return ret
}

func TrimDelimFrom(keepIndex int, key, delimiter string) string {
	split := strings.Split(key, delimiter)
	return strings.Join(split[keepIndex:], delimiter)
}

func (c *RedisClient) GetFloat(ctx context.Context, key string) (float64, error) {
	val := c.Get(ctx, key).Val()
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return floatVal, nil
}

func (c *RedisClient) Load(ctx context.Context, matrix string) (pipbot.FluidLevelMap, error) {
	ret := make(pipbot.FluidLevelMap)
	for key := range c.AllKeys(ctx, matrix) {
		val, err := c.GetFloat(ctx, key)
		if err != nil {
			return nil, err
		}
		ret[TrimDelimFrom(1, key, c.delimiter)] = val
	}
	return ret, nil
}

func (c *RedisClient) Flush(ctx context.Context, matrix string, m pipbot.FluidLevelMap) error {
	for k := range m {
		c.Set(ctx, c.Key(matrix, k), m[k], 0)
	}
	return nil
}
