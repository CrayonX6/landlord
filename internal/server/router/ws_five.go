package router

import (
    "beginning/internal/pkg/acme"
    "beginning/internal/server/http/behavior"
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "fmt"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "math"
    "sync"
)

var RWsFive WsFive

type WsFive struct {
    acme.UserInfo
}

type Player struct {
    Nickname   string          `json:"nickname"`    // 玩家昵称
    PlayerId   uint64          `json:"player_id"`   // 玩家ID，全场唯一
    RoomIndex  int             `json:"room_index"`  // 房间号
    Session    *melody.Session `json:"-"`           // 用户会话标识
    ChessColor string          `json:"chess_color"` // 棋色
    ChessStep  int             `json:"chess_step"`  // N手
}

// WebSocket five in a row
func (w *WsFive) WsFiveRoute(e *echo.Echo, path string) *melody.Melody {
    wsm := melody.New()

    // 所有玩家，通过会话找用户ID
    var sessionPlayers = make(map[*melody.Session]Player, 0)

    // 所有玩家，通过房间号找用户集合
    var roomPlayers []map[uint64]*melody.Session

    // 棋盘落子记录
    var chessHistory = make(map[int]map[int]string, 0)

    // 棋盘落子次数（顺序）
    var chessDown = make(map[int]int, 0)

    lock := new(sync.Mutex)
    e.GET(path, func(c echo.Context) error {
        _ = wsm.HandleRequest(c.Response().Writer, c.Request())
        return nil
    })

    colorMap := map[string]string{
        "black": "黑色方",
        "white": "白色方",
    }

    // 找房间坐下
    // 现有房间没有空位子则创建新房间
    var sitDown func(player *Player, times int64)
    sitDown = func(player *Player, times int64) {
        for roomIndex, items := range roomPlayers {
            if len(items) >= 2 {
                continue
            }
            player.ChessColor = "black"
            for _, item := range items {
                if sessionPlayers[item].ChessColor == "black" {
                    player.ChessColor = "white"
                } else {
                    player.ChessColor = "black"
                }
            }
            player.RoomIndex = roomIndex
            roomPlayers[roomIndex][player.PlayerId] = player.Session
            break
        }
        if player.RoomIndex < 0 {
            roomPlayers = append(roomPlayers, map[uint64]*melody.Session{})
            sitDown(player, times+1)
        }
    }

    // 返回指定房间的所有会话玩家
    var getSessionsByRoomIndex = func(roomIndex int) (players []*melody.Session) {
        for _, member := range roomPlayers[roomIndex] {
            players = append(players, member)
        }
        return
    }

    // 返回指定房间的所有玩家
    var getPlayerByRoomIndex = func(roomIndex int) (players []Player) {
        for _, member := range roomPlayers[roomIndex] {
            players = append(players, sessionPlayers[member])
        }
        return
    }

    // 返回下一个行动的棋色
    var getNextColorByRoomIndex = func(roomIndex int) string {
        if _, ok := chessDown[roomIndex]; !ok {
            chessDown[roomIndex] = 0
        }
        if chessDown[roomIndex]%2 == 0 {
            return "black"
        }
        return "white"
    }

    // 获取另外一个颜色
    var getOtherColor = func(color string) string {
        if color == "black" {
            return "white"
        }
        return "black"
    }

    // 计算是否胜出
    var calIsWin = func(roomIndex int, down int, color string) string {
        row := int(math.Floor(float64(down / 15)))
        col := down%15 - 1
        var getColor = func(d int) string {
            color := ""
            if _, ok := chessHistory[roomIndex][d]; ok {
                color = chessHistory[roomIndex][d]
            }
            return color
        }
        A := 1
        for i := 1; i <= 5; i++ { // 横向(-).左边
            if col-i < 0 || getColor(down-i) != color {
                break
            }
            A += 1
            if A >= 5 {
                return color
            }
        }
        for i := 1; i <= 5; i++ { // 横向(-).右边
            if col+i > 14 || getColor(down+i) != color {
                break
            }
            A += 1
            if A >= 5 {
                return color
            }
        }

        B := 1
        for i := 1; i <= 5; i++ { // 竖向(|).上边
            if row-i < 0 || getColor(down-(i*15)) != color {
                break
            }
            B += 1
            if B >= 5 {
                return color
            }
        }
        for i := 1; i <= 5; i++ { // 竖向(|).下边
            if row+i > 14 || getColor(down+(i*15)) != color {
                break
            }
            B += 1
            if B >= 5 {
                return color
            }
        }

        C := 1
        for i := 1; i <= 5; i++ { // 捺向(\).左上边
            if col-i < 0 || row-i < 0 || getColor(down-(i*15)-i) != color {
                break
            }
            C += 1
            if C >= 5 {
                return color
            }
        }
        for i := 1; i <= 5; i++ { // 捺向(\).右下边
            if col+i > 14 || row+i > 14 || getColor(down+(i*15)+i) != color {
                break
            }
            C += 1
            if C >= 5 {
                return color
            }
        }

        D := 1
        for i := 1; i <= 5; i++ { // 撇向(/).左下边
            if col-i < 0 || row+i > 14 || getColor(down+(i*15)-i) != color {
                break
            }
            D += 1
            if D >= 5 {
                return color
            }
        }
        for i := 1; i <= 5; i++ { // 撇向(/).右上边
            if col+i > 14 || row-i < 0 || getColor(down-(i*15)+i) != color {
                break
            }
            D += 1
            if D >= 5 {
                return color
            }
        }

        return ""
    }

    // 处理连接
    wsm.HandleConnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        id, _ := variables.Snowflake.NextID() // 给当前会话生成一个ID
        sessionPlayers[s] = Player{
            Nickname:  tool.ToStr(id),
            PlayerId:  id,
            RoomIndex: -1,
            Session:   s,
        }
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        me := sessionPlayers[s]
        // 对该房间的人进行广播.玩家下线
        msg := behavior.CBehavior.Message("offline", "", map[string]interface{}{
            "message": fmt.Sprintf("[%s] %s 下线了", colorMap[me.ChessColor], me.Nickname),
            "online":  len(roomPlayers[me.RoomIndex]) - 1,
        })
        _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(me.RoomIndex))
        delete(roomPlayers[me.RoomIndex], me.PlayerId)
        delete(sessionPlayers, s)
        chessHistory[me.RoomIndex] = make(map[int]string, 0)
        chessDown[me.RoomIndex] = 0
        for _, player := range getPlayerByRoomIndex(me.RoomIndex) {
            p := sessionPlayers[player.Session]
            p.ChessStep = 0
            sessionPlayers[player.Session] = p
        }
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        if args.Behavior == "register" {
            me := sessionPlayers[s]
            me.Nickname = args.Arguments["nickname"].(string)
            args.Arguments["user_id"] = int64(me.PlayerId)

            // 找有空位的房间坐下
            sitDown(&me, 1)
            sessionPlayers[s] = me

            // 返回注册信息
            args.Arguments["token"] = w.TokenCreator(args.Arguments)
            args.Arguments["room_number"] = me.RoomIndex + 1001
            behavior.CBehavior.WsResponse(s, args.Behavior, args.BehaviorId, args.Arguments)

            // 对该房间的人进行广播.玩家上线
            if _, ok := chessHistory[me.RoomIndex]; !ok {
                chessHistory[me.RoomIndex] = make(map[int]string)
            }
            msg := behavior.CBehavior.Message("online", args.BehaviorId, map[string]interface{}{
                "message":   fmt.Sprintf("[%s] %s 上线了", colorMap[me.ChessColor], me.Nickname),
                "online":    len(roomPlayers[me.RoomIndex]),
                "players":   getPlayerByRoomIndex(me.RoomIndex),
                "history":   chessHistory[me.RoomIndex],
                "nextColor": getNextColorByRoomIndex(me.RoomIndex),
            })
            _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(me.RoomIndex))
            return
        }

        _, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        if args.Behavior == "move" { // 移动棋子
            msg := behavior.CBehavior.Message(args.Behavior, args.BehaviorId, args.Arguments)
            _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(sessionPlayers[s].RoomIndex))
            return
        }

        if args.Behavior == "chessDown" { // 棋子落盘
            me := sessionPlayers[s]
            willDown := tool.ToInt(args.Arguments["willDown"])
            if _, ok := chessHistory[me.RoomIndex][willDown]; ok {
                return
            }

            me.ChessStep += 1
            sessionPlayers[s] = me
            args.Arguments["players"] = getPlayerByRoomIndex(me.RoomIndex)
            if _, ok := chessHistory[me.RoomIndex]; !ok {
                chessHistory[me.RoomIndex] = make(map[int]string, 0)
            }
            chessHistory[me.RoomIndex][willDown] = getNextColorByRoomIndex(me.RoomIndex)
            args.Arguments["nextColor"] = getOtherColor(chessHistory[me.RoomIndex][willDown])
            args.Arguments["history"] = chessHistory[me.RoomIndex]
            if _, ok := chessDown[me.RoomIndex]; !ok {
                chessDown[me.RoomIndex] = 1
            } else {
                chessDown[me.RoomIndex] += 1
            }

            args.Arguments["win"] = calIsWin(me.RoomIndex, willDown, chessHistory[me.RoomIndex][willDown])
            msg := behavior.CBehavior.Message(args.Behavior, args.BehaviorId, args.Arguments)
            _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(sessionPlayers[s].RoomIndex))
            return
        }

        return
    })

    return wsm
}
