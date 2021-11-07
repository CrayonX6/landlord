package landlords

import (
    "beginning/internal/entity"
    "beginning/internal/pkg/acme"
    "beginning/internal/server/http/behavior"
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "beginning/pkg/db"
    "fmt"
    "gopkg.in/olahol/melody.v1"
    "strings"
    "time"
)

var VCLandlords CLandlords

type CLandlords struct {
}

// 连接
func (c *CLandlords) Connect(wsm *melody.Melody, s *melody.Session) {
    /*
       var uid int64 = 10086
       uid2pokers[uid] = CreateBasicPoker(1)
       _, _ = ParsePokersMode(10086, []string{
           "3_mh", "3_fk", "3_hx", "3_ht",
           "4_mh", "4_fk", "4_hx", "9_ht",
           //"5_mh", "5_fk", "5_hx", "5_ht",
       }, "", 1, true)
    */
}

// 断开
func (c *CLandlords) DisConnect(wsm *melody.Melody, s *melody.Session) {
    uid := sid2uid[s]
    variables.Rds.LRem("ready_user", 0, uid)
    delete(uid2sid, uid)
    delete(sid2uid, s)
    delete(ready, uid)

    if rid, ok := uid2rid[uid]; ok { // 用户在对局中退出
        rid2surplus[rid] -= 1
        OnOffline(uid, "offline")
        if rid2surplus[rid] <= 0 { // 最后一人退出时销毁房间
            DestroyRoom(rid)
        }
    }
}

// 收到消息
func (c *CLandlords) Message(wsm *melody.Melody, s *melody.Session, args *behavior.WebSocketRequest, usr acme.UserInfoApi) {
    sid2uid[s] = usr.UID

    // 准备游戏
    if args.Behavior == "Ready" {
        uid2sid[usr.UID] = s
        if _, ok := uid2rid[usr.UID]; ok { // 该用户在对局中
            Comeback(usr.UID)
            return
        }

        u := &entity.User{}
        db.Orm.Where("id = ?", usr.UID).Find(&u)
        min := 1000
        if u.Bean < min {
            WsFailed(s, fmt.Sprintf("豆子＜%d无法开局", min), "ready")
            return
        }

        if _, ok := ready[usr.UID]; !ok {
            ready[usr.UID] = true
            variables.Rds.LPush("ready_user", usr.UID)
        }

        var rate = 5
        if _, ok := args.Arguments["rate"]; ok {
            rate = tool.ToInt(args.Arguments["rate"])
        }
        if rate < 5 {
            rate = 5
        }
        if rate > 95 {
            rate = 95
        }
        uid2rate[usr.UID] = rate

        // 创建房间
        var person int64 = 3
        if variables.Rds.LLen("ready_user").Val() >= person {
            _rid := rid
            rid += 1
            var i int64
            rid2uid[_rid] = []int64{}
            rate = 0
            for i = 1; i <= person; i++ {
                uid := int64(tool.ToInt(variables.Rds.RPop("ready_user").Val()))
                uid2rid[uid] = _rid
                rid2uid[_rid] = append(rid2uid[_rid], uid)
                rate += uid2rate[uid]
                delete(ready, usr.UID)
            }

            rid2uid[_rid] = ShuffleUser(rid2uid[_rid])
            rid2surplus[_rid] = person
            rid2grabNum[_rid] = 0
            rid2grabToken[_rid] = make(map[string]bool)
            rid2playingNum[_rid] = 0
            rid2playingToken[_rid] = make(map[string]bool)
            rid2playingLog[_rid] = make(map[int64]PlayingPokers)

            go func() {
                time.Sleep(time.Second * 7)
                if _, ok := rid2uid[_rid]; !ok {
                    return
                }
                GrabLords(wsm, s, usr, rid2grabNum[_rid])
            }()
            InitRoom(_rid, rate/3)
        }
    }

    // 取消准备
    if args.Behavior == "CancelReady" {
        variables.Rds.LRem("ready_user", 0, usr.UID)
        delete(uid2rate, usr.UID)
        delete(ready, usr.UID)
    }

    // 叫/抢地主
    if args.Behavior == "GrabLords" {
        receiveGrabLords(wsm, s, usr, args.Arguments, false)
    }

    // 玩家出牌
    if args.Behavior == "Playing" {
        receivePlaying(wsm, s, usr, args.Arguments)
    }
}

// 出牌
func receivePlaying(wsm *melody.Melody, s *melody.Session, usr acme.UserInfoApi, args map[string]interface{}) {
    var pokers []string
    var rid = uid2rid[usr.UID]
    tool.AlignStructAndMap(args["pokers"], &pokers)
    var shouldUid = GetNextPlayUid(rid)
    if shouldUid != usr.UID {
        WsFailed(s, "还未到你出牌", "")
        return
    }

    tool.PrintVar(strings.Repeat("-", 40))
    var playing PlayingPokers
    var prev = GetPrevKnockOut(rid)
    tool.PrintVar("上家牌", prev)

    if len(pokers) > 0 {
        var err error
        playing, err = ParsePokersMode(usr.UID, PokersUnique(pokers), prev.Mode, 1, false)
        if err != nil {
            WsFailed(s, err.Error(), "")
            return
        }
        // 对比上家出牌
        sameMode, err := ComparePlayingPokers(prev, playing)
        if err != nil {
            WsFailed(s, err.Error(), "")
            return
        }
        if sameMode && pokerMode[playing.Mode].DaNi {
            playing.Sound[1] = fmt.Sprintf("dani%d", tool.RandInt(1, 3))
        }
    } else {
        playing.Sound = []string{fmt.Sprintf("buyao%d", tool.RandInt(1, 4))}
    }

    // 删除用户牌
    var newUserPokers []Landlords
    for _, v := range uid2pokers[usr.UID] {
        if exists, _ := tool.InArray(v.Image, playing.Pokers); !exists {
            newUserPokers = append(newUserPokers, v)
        }
    }
    uid2pokers[usr.UID] = newUserPokers

    rid2playingNum[rid] += 1
    rid2playingLog[rid][rid2playingNum[rid]] = playing
    rid2playingToken[rid][args["token"].(string)] = true

    tool.PrintVar(fmt.Sprintf("用户%d打出了牌", usr.UID), playing)
    PlayingResult(wsm, usr.UID, playing)
    if len(newUserPokers) == 0 {
        Result(usr.UID)
    } else {
        time.Sleep(time.Second)
        Playing(wsm, rid)
    }
}

// 叫/抢地主
func receiveGrabLords(wsm *melody.Melody, s *melody.Session, usr acme.UserInfoApi, args map[string]interface{}, auto bool) {
    rid := uid2rid[usr.UID]
    if _, ok := rid2uid[rid]; !ok {
        return
    }
    var shouldUid = GetNextGrabLordUid(rid, rid2grabNum[rid])
    if !auto && shouldUid != usr.UID {
        WsFailed(s, "还未到你叫地主", "")
        return
    }
    if _, ok := args["token"]; !ok {
        return
    }
    rid2grabToken[rid][args["token"].(string)] = true

    // 广播动作
    GrabLordsResult(wsm, rid, rid2grabNum[rid], args)

    time.Sleep(time.Millisecond * 1500)
    rid2grabLog[rid] = append(rid2grabLog[rid], args["choice"].(bool))
    str := GrabLogJoinStr(rid)

    var l = len(rid2grabLog[rid])
    if l <= 2 {
        rid2grabNum[rid] += 1
        GrabLords(wsm, s, usr, rid2grabNum[rid])
    } else if l == 3 {
        if v, ok := map3[str]; ok {
            rid2lordUid[rid] = rid2uid[rid][v]
            DetermineLord(wsm, rid) // 确认地主
        } else {
            rid2grabNum[rid] += 1
            if str == "011" { // 特殊情况
                GrabLords(wsm, s, usr, rid2grabNum[rid]+1)
            } else {
                GrabLords(wsm, s, usr, rid2grabNum[rid])
            }
        }
    } else if l == 4 {
        rid2lordUid[rid] = rid2uid[rid][map4[str]]
        DetermineLord(wsm, rid) // 确认地主
    }
}
