package landlords

import (
    "beginning/internal/entity"
    "beginning/internal/server/http/behavior"
    tool "beginning/pkg/acme"
    "beginning/pkg/db"
    "errors"
    "fmt"
    "gopkg.in/olahol/melody.v1"
    "math/rand"
    "time"
)

// 响应失败
func WsFailed(s *melody.Session, message string, scene string) {
    behavior.CBehavior.WsResponse(s, "Failed", "", map[string]interface{}{
        "message": message,
        "scene":   scene,
    })
}

// 创建N副一例基础牌
func CreateBasicPk(n int, size int, char string) []Landlords {
    var pokers []Landlords
    type Design struct {
        Size int
        Char string
    }
    var design = []Design{
        {Size: 1, Char: "fk"},
        {Size: 2, Char: "mh"},
        {Size: 3, Char: "ht"},
        {Size: 4, Char: "hx"},
    }
    for i := 0; i < n; i++ {
        for _, d := range design {
            pokers = append(
                pokers,
                Landlords{
                    Image:      fmt.Sprintf("%s_%s", char, d.Char),
                    Design:     d.Char,
                    Size:       size*10 + d.Size,
                    NumSize:    size,
                    DesignSize: d.Size,
                    Char:       char,
                },
            )
        }
    }
    return pokers
}

// 创建N副基本牌
func CreateBasicPoker(n int) []Landlords {
    var pokers []Landlords
    type Deck struct {
        Size int
        Char string
    }
    var list = []Deck{
        {3, "3"},
        {4, "4"},
        {5, "5"},
        {6, "6"},
        {7, "7"},
        {8, "8"},
        {9, "9"},
        {10, "10"},
        {11, "j"},
        {12, "q"},
        {13, "k"},
        {14, "a"},
        {16, "2"},
    }
    for _, v := range list {
        pokers = append(pokers, CreateBasicPk(n, v.Size, v.Char)...)
    }

    return pokers
}

// 创建N副特殊牌
func CreateJokerPoker(n int) []Landlords {
    var pokers []Landlords
    type Deck struct {
        Size int
        Char string
    }
    var deckPoker = []Deck{
        {Size: 17, Char: "joker_small"},
        {Size: 18, Char: "joker_big"},
    }
    for i := 0; i < n; i++ {
        for _, p := range deckPoker {
            pokers = append(
                pokers,
                Landlords{
                    Image:   p.Char,
                    Size:    p.Size * 10,
                    NumSize: p.Size,
                    Char:    p.Char,
                },
            )
        }
    }
    return pokers
}

type Deck struct {
    Size int
    Char string
}

// 洗牌算法
func PokerAlgorithm(n int, mode string, customBoomRate int) (p [3][]Landlords, lord []Landlords) {
    var pokers []Landlords
    if mode == "custom" { // 不洗牌模式
        var list = ShufflePk([]Deck{
            {3, "3"},
            {4, "4"},
            {5, "5"},
            {6, "6"},
            {7, "7"},
            {8, "8"},
            {9, "9"},
            {10, "10"},
            {11, "j"},
            {12, "q"},
            {13, "k"},
            {14, "a"},
            {16, "2"},
        })

        // 基础牌
        var pokersChunk [][]Landlords
        for i := 0; i < n; i++ {
            for _, v := range list {
                boom := int(tool.RandInt(1, 100))
                pk := CreateBasicPk(1, v.Size, v.Char)
                if boom <= customBoomRate {
                    pokersChunk = append(pokersChunk, pk)
                } else {
                    for _, v := range pk {
                        pokersChunk = append(pokersChunk, []Landlords{v})
                    }
                }
            }
        }

        // 特殊牌
        for i := 0; i < n; i++ {
            boom := int(tool.RandInt(1, 100))
            pk := CreateJokerPoker(1)
            if boom <= customBoomRate {
                pokersChunk = append(pokersChunk, pk)
            } else {
                for _, v := range pk {
                    pokersChunk = append(pokersChunk, []Landlords{v})
                }
            }
        }

        // 打乱
        pokersChunk = ShufflePokersChunk(pokersChunk)
        for _, chunk := range pokersChunk {
            for _, v := range chunk {
                pokers = append(pokers, v)
            }
        }

    } else { // 普通模式
        pokers = append(CreateBasicPoker(n), CreateJokerPoker(n)...)
        pokers = ShufflePokers(pokers)
    }

    for k, v := range pokers {
        if k <= 50 {
            i := k / 17
            p[i] = append(p[i], v)
        } else {
            lord = append(lord, v)
        }
    }
    return
}

// 打乱例牌
func ShufflePk(pk []Deck) []Deck {
    if len(pk) <= 0 {
        return pk
    }
    rand.Seed(time.Now().Unix())
    for i := len(pk) - 1; i >= 0; i-- {
        num := rand.Intn(len(pk))
        pk[i], pk[num] = pk[num], pk[i]
    }
    return pk
}

// 打乱牌块顺序
func ShufflePokersChunk(chunk [][]Landlords) [][]Landlords {
    if len(chunk) <= 0 {
        return chunk
    }
    rand.Seed(time.Now().Unix())
    for i := len(chunk) - 1; i >= 0; i-- {
        num := rand.Intn(len(chunk))
        chunk[i], chunk[num] = chunk[num], chunk[i]
    }
    return chunk
}

// 打乱牌顺序
func ShufflePokers(pokers []Landlords) []Landlords {
    if len(pokers) <= 0 {
        return pokers
    }
    rand.Seed(time.Now().Unix())
    for i := len(pokers) - 1; i >= 0; i-- {
        num := rand.Intn(len(pokers))
        pokers[i], pokers[num] = pokers[num], pokers[i]
    }
    return pokers
}

// 打乱用户
func ShuffleUser(users []int64) []int64 {
    if len(users) <= 0 {
        return users
    }
    rand.Seed(time.Now().Unix())
    for i := len(users) - 1; i >= 0; i-- {
        num := rand.Intn(len(users))
        users[i], users[num] = users[num], users[i]
    }
    return users
}

// 叫地主的过程转为字符串
func GrabLogJoinStr(rid int64) string {
    var str string
    for _, v := range rid2grabLog[rid] {
        if v {
            str = fmt.Sprintf("%s%s", str, "1")
        } else {
            str = fmt.Sprintf("%s%s", str, "0")
        }
    }
    return str
}

// 对牌列表进行排序
func SortPokers(p []Landlords) []map[string]interface{} {
    var mm []map[string]interface{}
    var m []interface{}
    tool.AlignStructAndMap(p, &m)
    m = tool.SortArrayByFloat(m, "DESC", func(item interface{}) float64 {
        return item.(map[string]interface{})["size"].(float64)
    })
    tool.AlignStructAndMap(m, &mm)
    return mm
}

// 通过房间号查询玩家用户列表
func GetRoomUserInfo(rid int64) []UserInfo {
    var uList []UserInfo
    var skin = []string{"skin_zhuzhu", "skin_guoqilin", "skin_sichuan"}
    for k, u := range rid2uid[rid] {
        var t entity.User
        db.Orm.Where("id = ?", u).First(&t)
        var tt = UserInfo{
            Nickname: t.Nickname,
            Uid:      t.ID,
            Avatar:   t.HeadImg,
            Bean:     t.Bean,
            Pokers:   []map[string]interface{}{},
            Skin:     skin[k],
        }
        uList = append(uList, tt)
    }
    return uList
}

// 通过房间号获得 Session 数组
func getSessionListByRid(rid int64) []*melody.Session {
    var sList []*melody.Session
    for _, uid := range rid2uid[rid] {
        sid := uid2sid[uid]
        if sid != nil && !sid.IsClosed() {
            sList = append(sList, sid)
        }
    }
    return sList
}

// 出牌去重
func PokersUnique(pokers []string) []string {
    var m = map[string]bool{}
    var newPokers []string
    for _, v := range pokers {
        if _, ok := m[v]; !ok {
            newPokers = append(newPokers, v)
            m[v] = true
        }
    }
    return newPokers
}

// 字符串去重
func StringUnique(target []string) []string {
    var m = map[string]bool{}
    var newTarget []string
    for _, v := range target {
        if _, ok := m[v]; !ok {
            newTarget = append(newTarget, v)
            m[v] = true
        }
    }
    return newTarget
}

// 销毁房间
func DestroyRoom(rid int64) {
    for _, u := range rid2uid[rid] {
        delete(uid2rid, u)
        delete(uid2pokers, u)
    }
    delete(rid2uid, rid)
    delete(rid2surplus, rid)
    delete(rid2lordPokers, rid)
    delete(rid2lordUid, rid)
    delete(rid2grabNum, rid)
    delete(rid2grabLog, rid)
    delete(rid2grabToken, rid)
    delete(rid2playingNum, rid)
    delete(rid2playingLog, rid)
    delete(rid2playingToken, rid)
}

// 判断是否必须出牌
func IsMustKnockOut(rid int64) bool {
    var isMust = false
    var n = rid2playingNum[rid]
    if n == 0 {
        isMust = true
    } else {
        a, prev1 := rid2playingLog[rid][n]
        b, prev2 := rid2playingLog[rid][n-1]
        if prev1 && len(a.Pokers) == 0 && prev2 && len(b.Pokers) == 0 {
            isMust = true
        }
    }
    return isMust
}

// 获取上家/上上家出牌
func GetPrevKnockOut(rid int64) PlayingPokers {
    if IsMustKnockOut(rid) {
        return PlayingPokers{}
    }
    var n = rid2playingNum[rid]
    p, ok := rid2playingLog[rid][n]
    if ok && len(p.Pokers) > 0 {
        return p
    }
    return rid2playingLog[rid][n-1]
}

// 获取下一个出牌的人
func GetNextPlayUid(rid int64) int64 {
    var lordIndex int64 = -1
    for k, v := range rid2uid[rid] {
        if v == rid2lordUid[rid] {
            lordIndex = int64(k)
            break
        }
    }
    n := rid2playingNum[rid]
    index := (n + lordIndex) % int64(len(rid2uid[rid]))

    return rid2uid[rid][index]
}

// 获取下一个叫地主的人
func GetNextGrabLordUid(rid int64, playIndex int64) int64 {
    var index = playIndex % int64(len(rid2uid[rid]))
    return rid2uid[rid][index]
}

// 数组倒序函数
func Reverse(arr *[]string) {
    var temp string
    length := len(*arr)
    for i := 0; i < length/2; i++ {
        temp = (*arr)[i]
        (*arr)[i] = (*arr)[length-1-i]
        (*arr)[length-1-i] = temp
    }
}

// 用户在线/离开
func UserOnOffLine(uid int64) string {
    if _, ok := uid2sid[uid]; ok {
        return "在线"
    }
    return "离开"
}

// 三元运算.string
func TernaryStr(result bool, trueResult string, falseResult string) string {
    if result {
        return trueResult
    }
    return falseResult
}

// 和上家出牌对比
func ComparePlayingPokers(prev PlayingPokers, current PlayingPokers) (sameMode bool, err error) {
    if len(prev.Pokers) == 0 {
        return false, nil
    }
    if current.ModeSize > prev.ModeSize { // 牌型大于
        return false, nil
    }
    if current.ModeSize < prev.ModeSize { // 牌型小于
        return false, errors.New("牌型大不过")
    }
    if current.Mode != prev.Mode {
        return true, errors.New("牌型大不过")
    }
    if current.NumSize <= prev.NumSize {
        return true, errors.New("牌值大不过")
    }
    return true, nil
}
