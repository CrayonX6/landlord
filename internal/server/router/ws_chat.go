package router

import (
    "beginning/internal/pkg/acme"
    "beginning/internal/server/http/behavior"
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "beginning/pkg/handler"
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "net/http"
    "sync"
)

var RWsChat WsChat

type WsChat struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

// WebSocket chat
func (w *WsChat) WsChatRoute(e *echo.Echo, path string) *melody.Melody {
    wsm := melody.New()
    gophers := make(map[*melody.Session]string, 0)
    lock := new(sync.Mutex)
    onlineUser := 0
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
        gophers[s] = tool.ToStr(id)
        onlineUser += 1
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        msg := behavior.CBehavior.Message("offline", "", echo.Map{
            "nickname": gophers[s],
            "online":   onlineUser - 1,
        })
        _ = wsm.Broadcast([]byte(msg))
        delete(gophers, s)
        onlineUser -= 1
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        if args.Behavior == "register" {
            args.Arguments["user_id"] = int64(tool.ToInt(gophers[s]))
            if _, ok := args.Arguments["nickname"]; !ok {
                args.Arguments["nickname"] = gophers[s]
            }
            gophers[s] = args.Arguments["nickname"].(string)

            // 返回注册信息
            args.Arguments["token"] = w.TokenCreator(args.Arguments)
            behavior.CBehavior.WsResponse(s, args.Behavior, args.BehaviorId, args.Arguments)
            // 广播用户上线
            args.Arguments["online"] = onlineUser
            msg := behavior.CBehavior.Message("online", args.BehaviorId, args.Arguments)
            _ = wsm.Broadcast([]byte(msg))
            return
        }

        usr, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        go func() { // 单用户指令
            if args.Behavior == "chat" {
                args.Arguments["from"] = usr.Nickname
                args.Arguments["from_id"] = usr.UID
                args.Arguments["time"] = carbon.CreateFromTimestamp(args.Time).ToFormatString("m/d H:i:s")
                msg := behavior.CBehavior.Message(args.Behavior, args.BehaviorId, args.Arguments)
                _ = wsm.Broadcast([]byte(msg))
            }
        }()
        return
    })

    return wsm
}
