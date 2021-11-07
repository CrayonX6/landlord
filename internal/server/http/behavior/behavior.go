package behavior

import (
    "beginning/internal/pkg/acme"
    tool "beginning/pkg/acme"
    "beginning/pkg/handler"
    logger "beginning/pkg/log"
    "fmt"
    "github.com/golang-module/carbon"
    "gopkg.in/olahol/melody.v1"
)

var CBehavior Behavior

type Behavior struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

// 请求结构体
type WebSocketRequest struct {
    Authorization string `json:"Authorization"`
    acme.WebSocketResponse
}

// 转换为响应字符串
func (cc *Behavior) Message(behavior string, behaviorId string, arguments map[string]interface{}) string {
    return tool.InterfaceToJson(&acme.WebSocketResponse{
        Behavior:   behavior,
        BehaviorId: behaviorId,
        Arguments:  arguments,
        Time:       carbon.Now().ToTimestampWithSecond(),
    })
}

// 响应结构
func (cc *Behavior) WsResponse(s *melody.Session, behavior string, behaviorId string, arguments map[string]interface{}) {
    if s == nil || s.IsClosed() {
        return
    }
    response := cc.Message(behavior, behaviorId, arguments)
    err := s.Write([]byte(response))
    if err != nil {
        logger.Logger.Error(fmt.Sprintf("WebSocket发送消息出现错误：%s", err.Error()))
    }
}

// 响应成功
func (cc *Behavior) WsSuccess(s *melody.Session, message string, behaviorId string) {
    cc.WsResponse(s, "notice:success", behaviorId, map[string]interface{}{"message": message})
}

// 响应失败
func (cc *Behavior) WsFailed(s *melody.Session, message string, behaviorId string) {
    cc.WsResponse(s, "notice:failed", behaviorId, map[string]interface{}{"message": message})
}
