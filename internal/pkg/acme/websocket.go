package acme

import (
	"beginning/internal/entity"
	"beginning/internal/server/variables"
	tool "beginning/pkg/acme"
	"github.com/golang-module/carbon"
)

// WebSocket 响应结构体
type WebSocketResponse struct {
	Behavior   string                 `json:"Behavior"`
	BehaviorId string                 `json:"BehaviorId"`
	Arguments  map[string]interface{} `json:"Arguments"`
	Time       int64                  `json:"Time"`
}

// 实时推送用户行为（redis.List）
func (a *Acme) PushUserRealTimeBehavior(userId int64, wsr WebSocketResponse) int64 {
	key := entity.CK4UserRealTimeBehavior(userId)
	wsr.Time = carbon.Now().ToTimestampWithSecond()
	result := variables.Rds.LPush(key, tool.InterfaceToJson(wsr))
	return result.Val()
}

// 实时推送全服行为(redis.List)
func (a *Acme) PushAllRealTimeBehavior(wsr WebSocketResponse) int64 {
	key := entity.CK4AllRealTimeBehavior()
	wsr.Time = carbon.Now().ToTimestampWithSecond()
	result := variables.Rds.LPush(key, tool.InterfaceToJson(wsr))
	return result.Val()
}
