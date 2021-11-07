package router

import (
    "beginning/internal/entity"
    "beginning/internal/pkg/acme"
    "beginning/internal/server/http/behavior"
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "beginning/pkg/handler"
    "fmt"
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "net/http"
    "sync"
    "time"
)

var RWsBehavior WsBehavior
var BehaviorDebug = false

type WsBehavior struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

type BehaviorGophers struct {
    OnlineTime int64
    Nickname   string
}

// WebSocket behavior
func (w *WsBehavior) WsBehaviorRouter(e *echo.Echo, path string) *melody.Melody {
    wsm := melody.New()
    gophers := make(map[*melody.Session]BehaviorGophers, 0)
    lock := new(sync.Mutex)
    counter := 0
    wsm.Upgrader.CheckOrigin = func(r *http.Request) bool {
        return true
    }
    e.GET(path, func(c echo.Context) error {
        _ = wsm.HandleRequest(c.Response().Writer, c.Request())
        return nil
    })

    // 处理连接
    wsm.HandleConnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        id, _ := variables.Snowflake.NextID()
        gophers[s] = BehaviorGophers{
            OnlineTime: carbon.Now().ToTimestampWithSecond(),
            Nickname:   tool.ToStr(id),
        }
        counter += 1
        if BehaviorDebug {
            tool.PrintVar(fmt.Sprintf("客户端连接，连接后的用户数为：%d", counter))
        }
        if counter == 1 { // 全服行为监听
            go func() {
                for {
                    allBehavior := variables.Rds.BRPop(time.Minute, entity.CK4AllRealTimeBehavior()).Val()
                    if len(allBehavior) >= 2 {
                        arguments := &acme.WebSocketResponse{}
                        tool.JsonToInterface(allBehavior[1], &arguments)
                        if BehaviorDebug {
                            tool.PrintVar(fmt.Sprintf("收到全服队列消息：%+v", arguments))
                        }

                        // 不广播历史消息
                        t := carbon.Now().ToTimestampWithSecond()
                        if v, ok := gophers[s]; ok {
                            t = v.OnlineTime
                        }
                        if arguments.Time >= t {
                            msg := behavior.CBehavior.Message(arguments.Behavior, arguments.BehaviorId, arguments.Arguments)
                            _ = wsm.Broadcast([]byte(msg))
                        }
                    }
                    if counter <= 0 {
                        return
                    }
                }
            }()
        }
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        delete(gophers, s)
        counter -= 1
        if BehaviorDebug {
            tool.PrintVar(fmt.Sprintf("客户端断开连接，断开后的用户数为：%d", counter))
        }
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        if BehaviorDebug {
            tool.PrintVar(fmt.Sprintf("收到消息：%+v", args))
        }
        if args.Behavior == "ping" {
            behavior.CBehavior.WsResponse(s, "pong", args.BehaviorId, nil)
            return
        }

        usr, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        if args.Behavior == "behavior:userRealTime" { // 单用户行为监听
            go func() {
                for {
                    userBehavior := variables.Rds.BRPop(time.Minute, entity.CK4UserRealTimeBehavior(usr.UID)).Val()
                    if len(userBehavior) >= 2 {
                        arguments := &acme.WebSocketResponse{}
                        tool.JsonToInterface(userBehavior[1], &arguments)
                        if BehaviorDebug {
                            tool.PrintVar(fmt.Sprintf("收到单用户队列消息：%+v", arguments))
                        }
                        behavior.CBehavior.WsResponse(s, arguments.Behavior, arguments.BehaviorId, arguments.Arguments)
                    }
                    if t, ok := gophers[s]; ok && t.OnlineTime > 0 {
                        return
                    }
                }
            }()
            return
        }

        go func() { // 单用户指令
        }()
        return
    })

    return wsm
}
