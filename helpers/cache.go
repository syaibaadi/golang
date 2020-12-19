package helpers

import (
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.com/pangestu18/janji-online/chat/config"
	"gitlab.com/pangestu18/janji-online/chat/constant"
)

func CacheConnection() *redis.Client {
	zochc := redis.NewClient(&redis.Options{
		Addr:     config.Get("REDIS_HOST").String() + ":" + config.Get("REDIS_PORT").String(),
		Password: config.Get("REDIS_PASSWORD").String(),
		DB:       config.Get("REDIS_CACHE_DB").Int(),
	})
	return zochc
}

func QueueConnection() *redis.Client {
	zoqu := redis.NewClient(&redis.Options{
		Addr:     config.Get("REDIS_HOST").String() + ":" + config.Get("REDIS_PORT").String(),
		Password: config.Get("REDIS_PASSWORD").String(),
		DB:       config.Get("REDIS_DB").Int(),
	})
	return zoqu
}

func GetCache(ctx Context) *redis.Client {
	return ctx.Get(constant.CtxChc).(*redis.Client)
}

func GetCacheValue(ctx Context, key string) UnserializedPhpSession {
	key = config.Get("GO_CACHE_PREFIX").String() + ":" + key
	v, err := GetCache(ctx).Get(key).Result()
	if err != nil {
		return UnserializedPhpSession{Val: ""}
	} else {
		r, e := PhpUnserialize(v)
		if e == nil {
			return r
		} else {
			return UnserializedPhpSession{Val: ""}
		}
	}
}

func SetCacheValue(ctx Context, key string, val interface{}, expiration time.Duration) (bool, error) {
	key = config.Get("GO_CACHE_PREFIX").String() + ":" + key
	v, err := PhpSerialize(val)
	if err != nil {
		return false, err
	} else {
		err := GetCache(ctx).Set(key, v, expiration).Err()
		if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}

}
