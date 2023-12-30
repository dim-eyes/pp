package redis

import (
	"context"
	"math/rand"
	"pp/log"
	"runtime"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const RedisTypePlayer = 1 // 玩家缓存数据

var (
	redisMgr *RedisClientMgr
	logger   = log.GetLogger()
	Nil      = redis.Nil
)

// RedisClientMgr Redis管理类，用于获取各种类型的redis
// 业务链接的redis类型定义
// 使用方法： testMgr := GetInstance()
// 			testMgr.GetRedisClientByType(1) or testMgr.GetRedisClient()

type RedisClientMgr struct {
	redisClientMap map[int][]*RedisClient // map的key是redis的类型， value是RedisClient对象
}

func GetInstance() *RedisClientMgr {
	if redisMgr == nil {
		redisMgr = &RedisClientMgr{redisClientMap: make(map[int][]*RedisClient)}
		logger.Debug("instance RedisClient...")
	}
	return redisMgr
}

// AddRedisClientByType 根据redisType添加一个RedisClient
func (clientMgr *RedisClientMgr) AddRedisClientByType(redisType int, client *RedisClient) {
	clientMgr.redisClientMap[redisType] = append(clientMgr.redisClientMap[redisType], client)
}

// GetRedisClientByType 根据redisType随机获取一个RedisClient
func (clientMgr *RedisClientMgr) GetRedisClientByType(redisType int) (pClient *RedisClient, index int) {
	redisClients, ok := clientMgr.redisClientMap[redisType]
	if !ok {
		logger.Error("Get redis client error,redisType:", redisType)
		return nil, 0
	}

	redisClientCount := len(redisClients)
	if redisClientCount == 0 {
		logger.Error("Get redis client error,redisType:", redisType)
		return nil, 0
	} else if redisClientCount == 1 {
		return redisClients[0], 1
	}

	randIndex := rand.Intn(redisClientCount)
	pClient = redisClients[randIndex]
	return pClient, randIndex + 1
}

// AddRedisClient 添加一个Redisclient
func (clientMgr *RedisClientMgr) AddRedisClient(client *RedisClient) {
	clientMgr.redisClientMap[1] = append(clientMgr.redisClientMap[0], client)
}

// GetRedisClient 随机获取一个RedisClient
func (clientMgr *RedisClientMgr) GetRedisClient() (pClient *RedisClient, index int) {
	clients, ok := clientMgr.redisClientMap[0]
	if !ok {
		return nil, 0
	}

	redisClientCount := len(clients)
	if redisClientCount == 0 {
		return nil, 0
	}
	redisClients := clientMgr.redisClientMap[0]

	randIndex := rand.Intn(redisClientCount)

	pClient = redisClients[randIndex]
	logger.Debug("redis client index:", randIndex+1)
	return pClient, randIndex + 1
}

type RedisClient struct {
	ConnString string // "192.168.103.150:6379"
	Password   string
	rdb        *redis.Client
	ctx        context.Context
}

func (r *RedisClient) ConnRedis() bool {
	r.ctx = context.Background()
	r.rdb = redis.NewClient(&redis.Options{Addr: r.ConnString, Password: r.Password, PoolSize: 12 * runtime.NumCPU()})
	_, err := r.rdb.Ping(r.ctx).Result()
	if err != nil {
		logger.Error("connection redis error")
		return false
	}
	logger.Info("connection redis success")
	return true
}

// Pipeline 新建管道
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.rdb.Pipeline()
}

func (r *RedisClient) CallLua(luaString string, keys []string, args ...interface{}) {
	script := redis.NewScript(luaString)
	ret := script.Run(r.ctx, r.rdb, keys, args)
	logger.Debug(keys, args, ret)
}

// 检查给定 key 是否存在
func (r *RedisClient) Exists(key ...string) int64 {
	result := r.rdb.Exists(r.ctx, key...)
	logger.Debug(result)
	return result.Val()
}

// 用于设置 key 的过期时间，key 过期后将不再可用。单位以秒计。
func (r *RedisClient) Expire(key string, expiration time.Duration) bool {
	result := r.rdb.Expire(r.ctx, key, expiration)
	logger.Debug(result)
	return result.Val()
}

// 设置key的过期时间(在某个时间点过期)
func (r *RedisClient) ExpireAt(key string, expireTime int64) bool {
	tm := time.Unix(expireTime, 0)
	result := r.rdb.ExpireAt(r.ctx, key, tm)
	logger.Debug(result)
	return result.Val()
}

// TTL 用于获取 key 的过期时间
func (r *RedisClient) TTL(key string) int {
	result := r.rdb.TTL(r.ctx, key)
	rr := result.Val().String()
	if rr == "-2ns" {
		return 0
	} else {
		res, _ := strconv.Atoi(rr)
		return res
	}
}

// 获取指定 key 的值。如果 key 不存在，返回 nil 。如果key 储存的值不是字符串类型，返回一个错误。
func (r *RedisClient) Get(key string) string {
	result := r.rdb.Get(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

// 设置给定 key 的值。如果 key 已经存储其他值， SET 就覆写旧值，且无视类型。
func (r *RedisClient) Set(key string, value string) string {
	result := r.rdb.Set(r.ctx, key, value, 0)
	logger.Debug(result)
	return result.Val()
}

// （SET if Not eXists） 命令在指定的 key 不存在时，为 key 设置指定的值。
func (r *RedisClient) SetNX(key string, value interface{}, expiration time.Duration) bool {
	result := r.rdb.SetNX(r.ctx, key, value, expiration)
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) SetXX(key string, value interface{}, expiration time.Duration) bool {
	result := r.rdb.SetNX(r.ctx, key, value, expiration)
	logger.Debug(result)
	return result.Val()
}

// 删除已存在的键。不存在的 key 会被忽略。
func (r *RedisClient) Del(keys ...string) int64 {
	result := r.rdb.Del(r.ctx, keys...)
	logger.Debug(result)
	return result.Val()
}

/*
key 中储存的数字加上指定的增量值。

如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCRBY 命令。

如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。

本操作的值限制在 64 位(bit)有符号数字表示之内。
*/
func (r *RedisClient) Incrby(key string, value int64) int64 {
	result := r.rdb.IncrBy(r.ctx, key, value)
	logger.Debug(result)
	return result.Val()
}

// 命令返回所有(一个或多个)给定 key 的值。 如果给定的 key 里面，有某个 key 不存在，那么这个 key 返回特殊值 nil 。
func (r *RedisClient) MGet(key ...string) []interface{} {
	result := r.rdb.MGet(r.ctx, key...)
	logger.Debug(result)
	return result.Val()
}

// MSet is like Set but accepts multiple values:
//   - MSet("key1", "value1", "key2", "value2")
//   - MSet([]string{"key1", "value1", "key2", "value2"})
//   - MSet(map[string]interface{}{"key1": "value1", "key2": "value2"})

// 同时设置一个或多个 key-value 对。
func (r *RedisClient) MSet(values ...interface{}) string {
	result := r.rdb.MSet(r.ctx, values...)
	logger.Debug(result)
	return result.Val()
}

// 返回哈希表中指定字段的值。
func (r *RedisClient) HGet(key, field string) (string, error) {
	result := r.rdb.HGet(r.ctx, key, field)
	logger.Debug(result)
	return result.Val(), result.Err()
}

func (r *RedisClient) HLen(key string) int64 {
	result := r.rdb.HLen(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

/*
令用于返回哈希表中，所有的字段和值。
在返回值里，紧跟每个字段名(field name)之后是字段的值(value)，所以返回值的长度是哈希表大小的两倍。
*/
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	result := r.rdb.HGetAll(r.ctx, key)
	logger.Debug(result)
	return result.Val(), result.Err()
}

/*
HIncrBy
令用于为哈希表中的字段值加上指定增量值。
增量也可以为负数，相当于对指定字段进行减法操作。
如果哈希表的 key 不存在，一个新的哈希表被创建并执行 HINCRBY 命令。
如果指定的字段不存在，那么在执行命令前，字段的值被初始化为 0 。
对一个储存字符串值的字段执行 HINCRBY 命令将造成一个错误。
本操作的值被限制在 64 位(bit)有符号数字表示之内。
*/
func (r *RedisClient) HIncrBy(key, field string, incr int64) int64 {
	result := r.rdb.HIncrBy(r.ctx, key, field, incr)
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) HIncrByFloat(key, field string, incr float64) float64 {
	result := r.rdb.HIncrByFloat(r.ctx, key, field, incr)
	logger.Debug(result)
	return result.Val()
}

/*
	命令用于返回哈希表中，一个或多个给定字段的值。

如果指定的字段不存在于哈希表，那么返回一个 nil 值
*/
func (r *RedisClient) HMGet(key string, fields ...string) ([]interface{}, error) {
	result := r.rdb.HMGet(r.ctx, key, fields...)
	logger.Debug(result)
	return result.Val(), result.Err()
}

// HSet
// HSet is like Set but accepts multiple values:
//   - HSet("key1", "value1", "key2", "value2")
//   - HSet([]string{"key1", "value1", "key2", "value2"})
//   - HSet(map[string]interface{}{"key1": "value1", "key2": "value2"})
func (r *RedisClient) HSet(key string, fieldValues ...interface{}) (int64, error) {
	result := r.rdb.HSet(r.ctx, key, fieldValues...)
	logger.Debug(result)
	return result.Val(), result.Err()
}

func (r *RedisClient) HSetNX(key string, field string, value interface{}) bool {
	result := r.rdb.HSetNX(r.ctx, key, field, value)
	logger.Debug(result)
	return result.Val()
}

// HMSet
// HMSet is like Set but accepts multiple values:
//   - HMSet("key1", "value1", "key2", "value2")
//   - HMSet([]string{"key1", "value1", "key2", "value2"})
//   - HMSet(map[string]interface{}{"key1": "value1", "key2": "value2"})
/*命令用于为哈希表中的字段赋值 。
如果哈希表不存在，一个新的哈希表被创建并进行 HSET 操作。
如果字段已经存在于哈希表中，旧值将被覆盖。*/
func (r *RedisClient) HMSet(key string, fieldValues ...interface{}) (bool, error) {
	result := r.rdb.HMSet(r.ctx, key, fieldValues...)
	logger.Debug(result)
	return result.Val(), result.Err()
}

// 用于删除哈希表 key 中的一个或多个指定字段，不存在的字段将被忽略。
func (r *RedisClient) HDel(key string, field ...string) int64 {
	result := r.rdb.HDel(r.ctx, key, field...)
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) HExists(key string, field string) bool {
	result := r.rdb.HExists(r.ctx, key, field)
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) LLen(key string) int64 {
	result := r.rdb.LLen(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

// 命令将一个或多个值插入到列表头部。 如果 key 不存在，一个空列表会被创建并执行 LPUSH 操作。 当 key 存在但不是列表类型时，返回一个错误。
func (r *RedisClient) LPush(key string, values ...interface{}) int64 {
	result := r.rdb.LPush(r.ctx, key, values...)
	logger.Debug(result)
	return result.Val()
}
func (r *RedisClient) LPop(key string) (string, error) {
	result := r.rdb.LPop(r.ctx, key)
	logger.Debug(result)
	return result.Val(), result.Err()
}
func (r *RedisClient) LIndex(key string, index int64) string {
	result := r.rdb.LIndex(r.ctx, key, index)
	logger.Debug(result)
	return result.Val()
}

/*
命令用于将一个或多个值插入到列表的尾部(最右边)。
如果列表不存在，一个空列表会被创建并执行 RPUSH 操作。 当列表存在但不是列表类型时，返回一个错误。
*/
func (r *RedisClient) RPush(key string, values ...interface{}) (int64, error) {
	result := r.rdb.RPush(r.ctx, key, values...)
	logger.Debug(result)
	return result.Val(), result.Err()
}

func (r *RedisClient) RPop(key string) string {
	result := r.rdb.RPop(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

//SADD
/*将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。
假如集合 key 不存在，则创建一个只包含添加的元素作成员的集合。
当集合 key 不是集合类型时，返回一个错误。*/
func (r *RedisClient) SAdd(key string, members ...interface{}) int64 {
	result := r.rdb.SAdd(r.ctx, key, members...)
	logger.Debug(result)
	return result.Val()
}

//SIsMember
/*判断成员元素是否是集合的成员。*/
func (r *RedisClient) SIsMember(key string, member interface{}) bool {
	result := r.rdb.SIsMember(r.ctx, key, member)
	logger.Debug(result)
	return result.Val()
}

// 命令返回集合中元素的数量。
func (r *RedisClient) SCard(key string) int64 {
	result := r.rdb.SCard(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

// SMembers
// 返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
// 数据类型为[]string
func (r *RedisClient) SMembersSlice(key string) []string {
	result := r.rdb.SMembers(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) LRange(key string, start int64, stop int64) []string {
	result := r.rdb.LRange(r.ctx, key, start, stop)
	logger.Debug(result)
	return result.Val()
}

// SMembers
// 返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
// 数据类型为[]int64
func (r *RedisClient) SMembersIntSlice(key string) []int64 {
	//
	result := r.rdb.SMembers(r.ctx, key)
	logger.Debug(result)
	ret := make([]int64, 0)
	for _, value := range result.Val() {
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			ret = append(ret, valueInt)
		} else {
			ret = append(ret, 0)
		}
	}
	return ret
}

func (r *RedisClient) SMembers(key string) ([]string, error) {
	result := r.rdb.SMembers(r.ctx, key)
	logger.Debug(result)
	return result.Result()
}

// SMembers map
// 返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
// 数据类型为map[string]struct{}
func (r *RedisClient) SMembersMap(key string) map[string]struct{} {
	result := r.rdb.SMembersMap(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

// sscan
// 代集合中键的元素，Sscan 继承自 Scan。
func (r *RedisClient) SScan(key string, cursor uint64, match string, count int64) ([]string, uint64) {
	result := r.rdb.SScan(r.ctx, key, cursor, match, count)
	logger.Debug(result)
	return result.Val()
}

//SRem
/*命令用于移除集合中的一个或多个成员元素，不存在的成员元素会被忽略。
当 key 不是集合类型，返回一个错误。*/
func (r *RedisClient) SRem(key string, members ...interface{}) int64 {
	result := r.rdb.SRem(r.ctx, key, members)
	logger.Debug(result)
	return result.Val()
}

// SRANDMEMBER key
func (r *RedisClient) SRandMember(key string) string {
	result := r.rdb.SRandMember(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

//SRANDMEMBER key count
/*返回集合中的一个随机元素。
从 Redis 2.6 版本开始， Srandmember 命令接受可选的 count 参数：
如果 count 为正数，且小于集合基数，那么命令返回一个包含 count 个元素的数组，数组中的元素各不相同。如果 count 大于等于集合基数，那么返回整个集合。
如果 count 为负数，那么命令返回一个数组，数组中的元素可能会重复出现多次，而数组的长度为 count 的绝对值。*/
func (r *RedisClient) SRandMemberN(key string, count int64) []string {
	result := r.rdb.SRandMemberN(r.ctx, key, count)
	logger.Debug(result)
	return result.Val()
}

// sPop 移除集合中的指定 key 的一个随机元素，移除后会返回移除的元素。
func (r *RedisClient) SPop(key string) string {
	result := r.rdb.SPop(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

//ZAdd
/*命令用于将一个或多个成员元素及其分数值加入到有序集当中。
如果某个成员已经是有序集的成员，那么更新这个成员的分数值，并通过重新插入这个成员元素，来保证该成员在正确的位置上。
分数值可以是整数值或双精度浮点数。
如果有序集合 key 不存在，则创建一个空的有序集并执行 ZADD 操作。
当 key 存在但不是有序集类型时，返回一个错误。*/
func (r *RedisClient) ZAdd(key string, score int64, member string) int64 {
	result := r.rdb.ZAdd(r.ctx, key, &redis.Z{Member: member, Score: float64(score)})
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) ZAddFloat64(key string, score float64, member string) int64 {
	result := r.rdb.ZAdd(r.ctx, key, &redis.Z{Member: member, Score: score})
	logger.Debug(result)
	return result.Val()
}
func (r *RedisClient) ZAddNX(key string, score int64, member string) (int64, error) {
	result := r.rdb.ZAddNX(r.ctx, key, &redis.Z{Member: member, Score: float64(score)})
	logger.Debug(result)
	return result.Result()
}

// ZADDGT 参考  https://redis.io/commands/ZADD 这个包未支持   自己实现
func (r *RedisClient) ZAddGT(key string, score int64, member string) int64 {
	originnalScore := r.rdb.ZScore(r.ctx, key, member)
	if score > int64(originnalScore.Val()) {
		result := r.rdb.ZAdd(r.ctx, key, &redis.Z{Member: member, Score: float64(score)})
		return result.Val()
	}
	return int64(originnalScore.Val())
}

//ZIncrBy
/*对有序集合中指定成员的分数加上增量 increment
可以通过传递一个负数值 increment ，让分数减去相应的值，比如 ZINCRBY key -5 member ，就是让 member 的 score 值减去 5 。
当 key 不存在，或分数不是 key 的成员时， ZINCRBY key increment member 等同于 ZADD key increment member 。
当 key 不是有序集类型时，返回一个错误。
分数值可以是整数值或双精度浮点数。*/
func (r *RedisClient) ZIncrBy(key string, increment int64, member string) int64 {
	result := r.rdb.ZIncrBy(r.ctx, key, float64(increment), member)
	logger.Debug(result)
	return int64(result.Val())
}

func (r *RedisClient) ZIncrByFloat(key string, incrment float64, member string) float64 {
	result := r.rdb.ZIncrBy(r.ctx, key, incrment, member)
	logger.Debug(result.Val())

	return result.Val()
}

// ZRank
// 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。
func (r *RedisClient) ZRank(key string, member string) int64 {
	result := r.rdb.ZRank(r.ctx, key, member)
	logger.Debug(result)
	return int64(result.Val())
}

// ZRevRank
// 返回有序集中指定成员的排名。其中有序集成员按分数值递减(从大到小)倒序排列。
func (r *RedisClient) ZRevRank(key string, member string) int64 {
	result := r.rdb.ZRevRank(r.ctx, key, member)
	logger.Debug(result)
	if result.Err() == redis.Nil {
		return -1
	}
	return int64(result.Val())
}

func (r *RedisClient) ZCard(key string) int64 {
	result := r.rdb.ZCard(r.ctx, key)
	logger.Debug(result)
	return result.Val()
}

// ZREVRANGE key start stop
// 根据分数值从小到大排序
func (r *RedisClient) ZRange(key string, start, stop int64) []string {
	result := r.rdb.ZRange(r.ctx, key, start, stop)
	logger.Debug(result)
	return result.Val()
}

// ZREVRANGE key start stop
// 根据分数值从大到小排序
func (r *RedisClient) ZRevRange(key string, start, stop int64) ([]string, error) {
	result := r.rdb.ZRevRange(r.ctx, key, start, stop)
	logger.Debug(result)
	return result.Val(), result.Err()
}

// ZREVRANGE key start stop [WITHSCORES]
/*返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
具有相同分数值的成员按字典序的逆序(reverse lexicographical order )排列。
除了成员按分数值递减的次序排列这一点外， ZREVRANGEBYSCORE 命令的其他方面和 ZRANGEBYSCORE 命令一样。*/
func (r *RedisClient) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	result := r.rdb.ZRevRangeWithScores(r.ctx, key, start, stop)
	logger.Debug(result)
	return result.Val(), result.Err()
}

func (r *RedisClient) ZRangeWithScores(key string, start, stop int64) []redis.Z {
	result := r.rdb.ZRangeWithScores(r.ctx, key, start, stop)
	logger.Debug(result)
	return result.Val()
}

// ZScore key member
// 返回有序集中，成员的分数值。 如果成员元素不是有序集 key 的成员，或 key 不存在，返回 nil 。
func (r *RedisClient) ZScore(key string, member string) int64 {
	result := r.rdb.ZScore(r.ctx, key, member)
	logger.Debug(result)
	return int64(result.Val())
}

func (r *RedisClient) ZExist(key, member string) bool {
	result := r.rdb.ZScore(r.ctx, key, member)
	logger.Debug(result)
	if result.Err() == nil {
		return true
	} else {
		return false
	}
}

//ZRem key member
/*命令用于移除有序集中的一个或多个成员，不存在的成员将被忽略。
当 key 存在但不是有序集类型时，返回一个错误。*/
func (r *RedisClient) ZRem(key string, member interface{}) int64 {
	result := r.rdb.ZRem(r.ctx, key, member)
	logger.Debug(result)
	return result.Val()
}

// ZCount key start stop
// 返回有序集合中从start到stop分数区间的元素个数
func (r *RedisClient) ZCount(key string, start, stop int64) int64 {
	result := r.rdb.ZCount(r.ctx, key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10))
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) ZReveRangeByScore(key string, min, max int, offset, count int64) []string {
	result := r.rdb.ZRevRangeByScore(r.ctx, key, &redis.ZRangeBy{Min: strconv.Itoa(min), Max: strconv.Itoa(max), Offset: offset, Count: count})
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) ZReveRangeByScoreWithScore(key string, min, max int, offset, count int64) []redis.Z {
	result := r.rdb.ZRevRangeByScoreWithScores(r.ctx, key, &redis.ZRangeBy{Min: strconv.Itoa(min), Max: strconv.Itoa(max), Offset: offset, Count: count})
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) ZReveRangeByScoreWithScoreByString(key string, min string, max string, offset, count int64) []redis.Z {
	result := r.rdb.ZRevRangeByScoreWithScores(r.ctx, key, &redis.ZRangeBy{Min: min, Max: max, Offset: offset, Count: count})
	logger.Debug(result)
	return result.Val()
}

func (r *RedisClient) ZRangeByScore(key string, min, max int64, offset, count int64) []string {
	result := r.rdb.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min:    strconv.FormatInt(min, 10),
		Max:    strconv.FormatInt(max, 10),
		Offset: offset,
		Count:  count,
	})
	logger.Debug(result)
	return result.Val()
}
