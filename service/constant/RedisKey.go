package constant

import "fmt"

const (
	// PlayerStatusRedisKey 玩家所在的网关数据，玩家是否离线数
	PlayerStatusRedisKey = "player:roomProfile:hash:%d"
)

// 相关redis key
// GetPlayerStatusKey 获取用户网关信息 redis key
func GetPlayerStatusKey(userId int) string {
	return fmt.Sprintf(PlayerStatusRedisKey, userId)
}
