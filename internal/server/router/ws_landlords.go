package router

import (
    "beginning/internal/pkg/acme"
    "beginning/internal/server/http/behavior"
    "beginning/internal/server/http/landlords"
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "beginning/pkg/handler"
    "fmt"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "net/http"
    "sync"
    "time"
)

var RWsLandlords WsLandlords

type WsLandlords struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

// WebSocket landlords
func (w *WsLandlords) WsLandlordsRoute(e *echo.Echo, path string) *melody.Melody {
    wsm := melody.New()
    lock := new(sync.Mutex)
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
        landlords.VCLandlords.Connect(wsm, s)
        lock.Unlock()
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        landlords.VCLandlords.DisConnect(wsm, s)
        lock.Unlock()
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        // 心跳
        if args.Behavior == "ping" {
            behavior.CBehavior.WsResponse(s, "pong", args.BehaviorId, nil)
            return
        }

        usr, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        key := fmt.Sprintf("rate_limit:%d:%s", usr.UID, args.Behavior)
        if variables.Rds.Exists(key).Val() {
            return
        }
        variables.Rds.Set(key, true, time.Millisecond*500)
        landlords.VCLandlords.Message(wsm, s, args, usr)
        return
    })

    return wsm
}
