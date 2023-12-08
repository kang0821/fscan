package client

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/shadow1ng/fscan/config"
	"strings"
	"time"
)

type RedisContext struct {
	ctx         context.Context
	RedisClient *redis.ClusterClient
}

func InitRedis(redisConfig config.Redis) {
	Context.Redis.ctx = context.Background()
	Context.Redis.RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    redisConfig.Nodes,
		Password: redisConfig.Password,
	})
}

func (redisContext *RedisContext) SGet(k string) []string {
	result := redisContext.RedisClient.SMembers(Context.Redis.ctx, k).Val()
	return convert(result)
}

func (redisContext *RedisContext) Del(keys ...string) int64 {
	return redisContext.RedisClient.Del(Context.Redis.ctx, keys...).Val()
}

func getExpire(ex interface{}) time.Duration {
	if ex == nil {
		return time.Duration(0) * time.Second
	}
	return time.Duration(ex.(int)) * time.Second
}

// go-redis获取的字符串是带双引号的，这里暂时直接替换掉  TODO
func convert(values []string) []string {
	var result []string
	for _, v := range values {
		if strings.Contains(v, "\"") {
			v = strings.ReplaceAll(v, "\"", "")
		}
		result = append(result, v)
	}
	return result
}
