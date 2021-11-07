package landlords

import (
    "beginning/internal/entity"
    "beginning/internal/pkg/acme"
    "beginning/internal/server/http/behavior"
    tool "beginning/pkg/acme"
    "beginning/pkg/db"
    "fmt"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "gorm.io/gorm"
    "time"
)

// 用户返回房间
func Comeback(uid int64) {
    var uList []UserInfo
    uList = GetRoomUserInfo(uid2rid[uid])
    sid := uid2sid[uid]
    pokers := SortPokers(uid2pokers[uid])

    rid = uid2rid[uid]
    rid2surplus[rid] += 1

    lordStr := TernaryStr(rid2lordUid[rid] > 0, "地主", "")
    farmerStr := TernaryStr(rid2lordUid[rid] > 0, "农民", "")
    for k, u := range uList {
        uList[k].Role = TernaryStr(u.Uid == rid2lordUid[rid], lordStr, farmerStr)
        if u.Uid != uid { // 不是自己的信息
            uList[k].Lord = rid2lordUid[rid] == u.Uid
            uList[k].PokersNum = fmt.Sprintf("%dP (%s)", len(uid2pokers[u.Uid]), UserOnOffLine(u.Uid))
        } else {
            uList[k].Lord = rid2lordUid[rid] == u.Uid
            uList[k].PokersNum = fmt.Sprintf("%dP (%s)", len(pokers), UserOnOffLine(u.Uid))
            uList[k].Pokers = pokers
        }
    }

    var rid = uid2rid[uid]
    var lordPoker []Landlords
    if _, ok := rid2lordUid[rid]; ok {
        lordPoker = rid2lordPokers[rid]
    }

    go func() {
        OnOffline(uid, "online")
    }()

    behavior.CBehavior.WsResponse(sid, "Comeback", "", map[string]interface{}{
        "users": uList,
        "lord":  lordPoker,
    })
}

// 用户上线/离线
func OnOffline(uid int64, action string) {
    var uList []UserInfo
    uList = GetRoomUserInfo(uid2rid[uid])
    rid = uid2rid[uid]

    lordStr := TernaryStr(rid2lordUid[rid] > 0, "地主", "")
    farmerStr := TernaryStr(rid2lordUid[rid] > 0, "农民", "")
    for _, uu := range rid2uid[rid] {
        if uu == uid {
            continue
        }
        for k, u := range uList {
            uList[k].Role = TernaryStr(u.Uid == rid2lordUid[rid], lordStr, farmerStr)
            uList[k].PokersNum = fmt.Sprintf("%dP (%s)", len(uid2pokers[u.Uid]), UserOnOffLine(u.Uid))
            if u.Uid != uu { // 不是自己的信息
                if u.Uid == rid2lordUid[rid] { // 是地主
                    uList[k].Lord = true
                }
            } else {
                uList[k].Pokers = SortPokers(uid2pokers[u.Uid])
                if u.Uid == rid2lordUid[rid] { // 是地主
                    uList[k].Lord = true
                }
            }
        }
        behavior.CBehavior.WsResponse(uid2sid[uu], "OnOffline", "", map[string]interface{}{
            "users": uList,
        })
    }
}

// 初始化房间
func InitRoom(rid int64, customBoomRate int) {
    p, lord := PokerAlgorithm(1, "custom", customBoomRate)
    rid2lordPokers[rid] = lord

    var uList = GetRoomUserInfo(rid)
    for k, uid := range rid2uid[rid] {
        uid2pokers[uid] = p[k]
        sid := uid2sid[uid]
        mm := SortPokers(uid2pokers[uid])

        for k, u := range uList {
            if u.Uid == uid {
                uList[k].Pokers = mm
            }
        }
        behavior.CBehavior.WsResponse(sid, "InitRoom", "", map[string]interface{}{
            "users": uList,
            "rate":  customBoomRate,
        })
    }
    return
}

// 抢地主
func GrabLords(wsm *melody.Melody, s *melody.Session, usr acme.UserInfoApi, playIndex int64) {
    var hasChoice = false
    var rid = uid2rid[usr.UID]
    if _, ok := rid2uid[rid]; !ok {
        return
    }
    for k := range rid2grabLog[rid] {
        if rid2grabLog[rid][k] == true {
            hasChoice = true
        }
    }

    var second = 15
    var token = tool.RandStr(32)

    msg := behavior.CBehavior.Message("GrabLords", "", echo.Map{
        "uid":        GetNextGrabLordUid(rid, playIndex),
        "has_choice": hasChoice,
        "second":     second,
        "token":      token,
    })

    go func() {
        time.Sleep(time.Second * time.Duration(second+1))
        if _, ok := rid2uid[rid]; !ok {
            return
        }
        if _, ok := rid2lordUid[rid]; ok {
            return
        }
        if _, ok := rid2grabToken[rid][token]; !ok { // 倒计时后还未做出选择
            var text string
            var sound string
            if hasChoice {
                text = "不抢"
                sound = "buqiang"
            } else {
                text = "不叫"
                sound = "bujiao"
            }
            receiveGrabLords(wsm, s, usr, echo.Map{
                "token":  token,
                "choice": false,
                "text":   text,
                "sound":  sound,
            }, true)
        }
    }()

    sList := getSessionListByRid(rid)
    if len(sList) > 0 {
        _ = wsm.BroadcastMultiple([]byte(msg), sList)
    }
}

// 广播叫地主的情况
func GrabLordsResult(wsm *melody.Melody, rid int64, playIndex int64, args map[string]interface{}) {
    msg := behavior.CBehavior.Message("GrabLordsResult", "", echo.Map{
        "uid":   GetNextGrabLordUid(rid, playIndex),
        "text":  args["text"],
        "sound": args["sound"],
    })

    sList := getSessionListByRid(rid)
    if len(sList) > 0 {
        _ = wsm.BroadcastMultiple([]byte(msg), sList)
    }
}

// 确定地主
func DetermineLord(wsm *melody.Melody, rid int64) {
    var uList = GetRoomUserInfo(rid)
    // 标识三张为地主牌
    var lordPokers = rid2lordPokers[rid]
    for k, _ := range lordPokers {
        lordPokers[k].Lord = true
    }

    // 修改对局玩家的牌（地主）
    uid2pokers[rid2lordUid[rid]] = append(uid2pokers[rid2lordUid[rid]], lordPokers...)

    for _, uid := range rid2uid[rid] {
        for k, u := range uList {
            uList[k].Role = TernaryStr(u.Uid == rid2lordUid[rid], "地主", "农民")
            uList[k].PokersNum = fmt.Sprintf("%dP (%s)", len(uid2pokers[u.Uid]), UserOnOffLine(u.Uid))
            if u.Uid != uid { // 不是自己的信息
                if u.Uid == rid2lordUid[rid] { // 是地主
                    uList[k].Lord = true
                }
            } else {
                uList[k].Pokers = SortPokers(uid2pokers[u.Uid])
                if u.Uid == rid2lordUid[rid] { // 是地主
                    uList[k].Lord = true
                }
            }
        }
        behavior.CBehavior.WsResponse(uid2sid[uid], "DetermineLord", "", map[string]interface{}{
            "users": uList,
            "lord":  rid2lordPokers[rid],
        })
    }

    go func() {
        time.Sleep(time.Second * 5)
        if _, ok := rid2uid[rid]; !ok {
            return
        }
        Playing(wsm, rid)
    }()
}

// 出牌
func Playing(wsm *melody.Melody, rid int64) {
    if _, ok := rid2uid[rid]; !ok {
        return
    }
    var second = 20
    var token = tool.RandStr(32)
    var uid = GetNextPlayUid(rid)

    msg := behavior.CBehavior.Message("Playing", "", echo.Map{
        "token":  token,
        "uid":    uid,
        "no":     !IsMustKnockOut(rid),
        "second": second,
    })

    sList := getSessionListByRid(rid)
    if len(sList) > 0 {
        _ = wsm.BroadcastMultiple([]byte(msg), sList)
    }
}

// 广播出牌的情况
func PlayingResult(wsm *melody.Melody, uid int64, playing PlayingPokers) {
    var rid = uid2rid[uid]
    if _, ok := rid2uid[rid]; !ok {
        return
    }
    var pks = playing.Pokers
    Reverse(&pks)
    if len(pks) == 0 {
        pks = []string{}
    }

    msg := behavior.CBehavior.Message("PlayingResult", "", echo.Map{
        "uid":    uid,
        "pokers": pks,
        "sound":  playing.Sound,
        "num":    fmt.Sprintf("%dP (%s)", len(uid2pokers[uid]), UserOnOffLine(uid)),
    })

    sList := getSessionListByRid(rid)
    if len(sList) > 0 {
        _ = wsm.BroadcastMultiple([]byte(msg), sList)
    }
}

// 广播对局结果
func Result(win int64) {
    rid = uid2rid[win]
    lordUid := rid2lordUid[rid]
    var winUser []int64
    var loseUser []int64

    var factor = 1000
    var nowBean = make(map[int64]int)
    var operateBean = make(map[int64]int)
    for _, v := range GetRoomUserInfo(rid) {
        nowBean[v.Uid] = v.Bean
    }

    if lordUid == win { // 地主赢了
        gotTotal := 0
        for _, v := range rid2uid[rid] {
            if v != lordUid {
                loseUser = append(loseUser, v)
                lose := factor
                if lose > nowBean[v] {
                    lose = nowBean[v]
                }
                gotTotal += lose
                operateBean[v] = lose * -1
            }
        }
        winUser = append(winUser, lordUid)
        operateBean[lordUid] = gotTotal

    } else { // 农民赢了

        loseUser = append(loseUser, lordUid)
        loseTotal := factor * 2
        if loseTotal > nowBean[lordUid] {
            loseTotal = nowBean[lordUid]
        }
        operateBean[lordUid] = loseTotal * -1
        for _, v := range rid2uid[rid] {
            if v != lordUid {
                winUser = append(winUser, v)
                operateBean[v] = loseTotal / (len(rid2uid[rid]) - 1)
            }
        }
    }

    // 更新操作
    for u, b := range operateBean {
        db.Orm.Model(&entity.User{}).Where("id = ?", u).Update("bean", gorm.Expr("bean + ?", b))
    }

    // 结算
    for _, u := range loseUser {
        behavior.CBehavior.WsResponse(uid2sid[u], "Result", "", map[string]interface{}{
            "result": "lose",
            "bean":   operateBean[u],
        })
    }
    for _, u := range winUser {
        behavior.CBehavior.WsResponse(uid2sid[u], "Result", "", map[string]interface{}{
            "result": "win",
            "bean":   operateBean[u],
        })
    }
    // 销毁房间
    DestroyRoom(rid)
}
