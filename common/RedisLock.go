package common

import (
	"crypto/rand"
	"encoding/base64"
	mathRand "math/rand"
	"pp/db/redis"
	"time"
)

type RedLock struct {
	Key   string
	Value string
}

const (
	retryMaxCount = 10
)

func (r *RedLock) LockTryOne(key string, ttl time.Duration) bool {
	client, index := redis.GetInstance().GetRedisClientByType(redis.RedisTypePlayer)
	if index == 0 {
		return false
	}
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return false
	}
	value := base64.StdEncoding.EncodeToString(b)
	lockResult := client.SetNX(key, value, ttl)
	if lockResult {
		r.Key = key
		r.Value = value
		return true
	} else {
		return false
	}
}

func (r *RedLock) Lock(key string, ttl time.Duration) bool {
	client, index := redis.GetInstance().GetRedisClientByType(redis.RedisTypePlayer)
	if index == 0 {
		return false
	}

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return false
	}
	value := base64.StdEncoding.EncodeToString(b)
	for retryCount := 0; retryCount < retryMaxCount; retryCount++ {
		lockResult := client.SetNX(key, value, ttl)
		if lockResult {
			r.Key = key
			r.Value = value
			return true
		}
		mi := mathRand.Int31n(300) + 50
		time.Sleep(time.Millisecond * time.Duration(mi))
	}
	return false
}

func (r *RedLock) Unlock() {
	redis, index := redis.GetInstance().GetRedisClientByType(redis.RedisTypePlayer)
	if index == 0 {
		return
	}
	redis.CallLua(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end`, []string{r.Key}, r.Value)
}
