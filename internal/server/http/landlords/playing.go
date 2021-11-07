package landlords

import (
    tool "beginning/pkg/acme"
    "errors"
    "fmt"
    "math"
    "strings"
)

// 解析打出牌的牌型
func ParsePokersMode(uid int64, pokers []string, hitMode string, times int, debug bool) (playing PlayingPokers, err error) {
    defer func() {
        if err == nil {
            playing.ModeSize = pokerMode[playing.Mode].Size
            var sound = playing.Sound
            playing.Sound = append(playing.Sound, pokerMode[playing.Mode].Sound...)
            playing.Sound = append(playing.Sound, sound...)
            if debug {
                tool.PrintVar(fmt.Sprintf("第%d次：断言牌型成功：%s", times, playing.Mode), playing)
            }
        }
        if len(playing.Pokers) == 0 {
            playing.Pokers = pokers
        }
    }()

    // 玩家手里的牌
    var userPoker = make(map[string]Landlords)
    for _, landlords := range uid2pokers[uid] {
        userPoker[landlords.Image] = landlords
    }
    if len(userPoker) == 0 {
        err = errors.New("玩家已经没牌了")
        return
    }

    var outPoker = make(map[string]map[string]interface{})
    var sortedPks []interface{}
    for _, img := range pokers {

        //  出的牌是否都真实存在
        if _, ok := userPoker[img]; !ok {
            err = errors.New("玩家出牌非法")
            return
        }

        // img牌 -> 结构体牌
        pk := userPoker[img]
        sortedPks = append(sortedPks, pk)

        // 对出牌数值进行分组
        numSizeStr := tool.ToStr(pk.NumSize)
        if _, ok := outPoker[numSizeStr]; !ok {
            outPoker[numSizeStr] = map[string]interface{}{
                "times": 1,
                "size":  pk.NumSize,
                "image": []Landlords{pk},
            }
        } else {
            outPoker[numSizeStr]["times"] = outPoker[numSizeStr]["times"].(int) + 1
            outPoker[numSizeStr]["image"] = append(outPoker[numSizeStr]["image"].([]Landlords), pk)
        }
    }

    // 对出的牌单纯从小到大排序
    sortedPks = tool.SortArrayByFloat(sortedPks, "ASC", func(item interface{}) float64 {
        return float64(item.(Landlords).NumSize)
    })

    // 判断简单牌型
    var sortedPokers []Landlords
    tool.AlignStructAndMap(sortedPks, &sortedPokers)
    var pokersLen = len(sortedPokers)

    // 单牌
    if pokersLen == 1 {
        playing.Mode = "1"
        playing.NumSize = sortedPokers[0].NumSize
        playing.Sound = []string{sortedPokers[0].Char}
        return
    }

    // 对子
    if pokersLen == 2 {
        if sortedPokers[0].Image == "joker_small" && sortedPokers[1].Image == "joker_big" {
            playing.Mode = "xy"
        } else if sortedPokers[0].NumSize == sortedPokers[1].NumSize {
            playing.Mode = "11"
            playing.NumSize = sortedPokers[0].NumSize
            playing.Sound = []string{strings.Repeat(sortedPokers[0].Char, pokersLen)}
        } else {
            if debug {
                tool.PrintVar(fmt.Sprintf("第%d次：玩家对子牌型非法", times), pokers)
            }
            err = errors.New("牌型不支持")
        }
        return
    }

    // 三张
    if pokersLen == 3 {
        if sortedPokers[0].NumSize == sortedPokers[1].NumSize && sortedPokers[0].NumSize == sortedPokers[2].NumSize {
            playing.Mode = "111"
            playing.NumSize = sortedPokers[0].NumSize
            playing.Sound = []string{strings.Repeat(sortedPokers[0].Char, pokersLen)}
        } else {
            if debug {
                tool.PrintVar(fmt.Sprintf("第%d次：玩家三张牌型非法", times), pokers)
            }
            err = errors.New("牌型不支持")
        }
        return
    }

    // 对出的牌综合排序，优先次数从大到小，其次数值从小到大
    for k, v := range outPoker {
        times := v["times"].(int)
        outPoker[k]["sort"] = (times * 1000) + (100 - v["size"].(int)) // 计算综合排序的字段值
    }

    y := tool.MapToArray(outPoker, "")
    var sortedOutPokers = tool.SortMapArray(y, "sort", "DESC")

    type Chunk struct {
        Sheet     int // 总牌张数
        LatestNum int
        Pokers    []Landlords
    }

    // 分出一堆（用于计算主牌），此过程中有部分牌可以直接断言为从牌
    var slave []Landlords
    var chunk = map[int]Chunk{}
    for _, v := range sortedOutPokers {
        times := v["times"].(int)
        size := v["size"].(int)
        image := v["image"].([]Landlords)
        if _, ok := chunk[times]; !ok {
            chunk[times] = Chunk{
                Sheet:     times,
                LatestNum: size,
                Pokers:    image,
            }
        } else {
            if chunk[times].LatestNum+1 == size {
                item := chunk[times]
                item.Sheet += times
                item.LatestNum = size
                item.Pokers = append(item.Pokers, image...)
                chunk[times] = item
            } else {
                slave = append(slave, image...)
            }
        }
    }

    // 计算主牌
    var masterKey = 0
    var masterSheet = chunk[masterKey].Sheet
    for k, v := range chunk {
        if v.Sheet > masterSheet {
            masterSheet = v.Sheet
            masterKey = k
        }
    }
    var master = chunk[masterKey].Pokers
    for k, v := range chunk {
        if k != masterKey {
            slave = append(slave, v.Pokers...)
        }
    }

    // 生成主牌牌型
    var masterMode string
    var minNumSize = master[0].NumSize
    for _, v := range master {
        masterMode = fmt.Sprintf("%s%d", masterMode, v.NumSize-minNumSize+1)
    }

    // 获取打出去的牌顺序
    var latestPokers = func() (latest []string) {
        for _, v := range master {
            latest = append(latest, v.Image)
        }
        for _, v := range slave {
            latest = append(latest, v.Image)
        }
        return
    }

    var dan = len(slave)
    var dui = 0
    for i := 0; i < len(slave); i++ {
        if len(slave) < i+2 {
            continue
        }
        if slave[i].NumSize == slave[i+1].NumSize {
            dui += 1
            i += 1
        }
    }

    // 特殊牌型（连炸牌型）
    var numSizeAdd = 0
    if dan == 0 {
        split := strings.Split(masterMode, "")
        split = StringUnique(split)
        if len(split)*4 == len(masterMode) {
            if m, ok := pokerMode[masterMode]; ok {
                if exists, _ := tool.InArray(hitMode, m.Same); exists {
                    masterMode = hitMode
                    if masterMode == "1111_2dui" {
                        numSizeAdd = 1
                    }
                } else if len(m.Same) > 0 {
                    masterMode = m.Same[0]
                }
            }
        }
    }

    if dan == 0 { // 不带任何从牌
        if _, ok := pokerMode[masterMode]; ok {
            playing.Mode = masterMode
            playing.NumSize = minNumSize + numSizeAdd
            playing.Pokers = latestPokers()
            return
        } else if debug {
            tool.PrintVar(fmt.Sprintf("第%d次：断言牌型：%s", times, masterMode))
        }

    } else {

        var modeDan = fmt.Sprintf("%s_%ddan", masterMode, dan)
        var modeDui = fmt.Sprintf("%s_%ddui", masterMode, dui)

        // 优先判断带对子从牌
        if dui*2 == dan { // 两对为炸弹拆分的
            if dui == 2 && slave[0].NumSize == slave[1].NumSize && slave[0].NumSize == slave[2].NumSize && slave[0].NumSize == slave[3].NumSize {
                numSizeAdd = slave[3].NumSize - minNumSize
            }
            if _, ok := pokerMode[modeDui]; ok {
                playing.Mode = modeDui
                playing.NumSize = minNumSize + numSizeAdd
                playing.Pokers = latestPokers()
                return
            } else if debug {
                tool.PrintVar(fmt.Sprintf("第%d次：断言牌型失败：%s", times, modeDui))
            }
        }

        // 其次判断带单牌从牌
        if _, ok := pokerMode[modeDan]; ok {
            playing.Mode = modeDan
            playing.NumSize = minNumSize + numSizeAdd
            playing.Pokers = latestPokers()
            return
        } else if debug {
            tool.PrintVar(fmt.Sprintf("第%d次：断言牌型失败：%s", times, modeDan))
        }
    }

    // 牌型不存在并且从牌比较多的情况
    if dan > 3 && dan > int(math.Floor(float64(len(master))*0.8)) {
        var slavePokers []string
        for _, v := range slave {
            slavePokers = append(slavePokers, v.Image)
        }
        // 二次解析从牌牌型
        pg, _ := ParsePokersMode(uid, slavePokers, "", times+1, debug)
        if pg.Mode == "111_1dan" && masterMode == "1111" && math.Abs(float64(minNumSize-pg.NumSize)) == 1 {
            playing.Mode = "111222_2dan"
            playing.NumSize = int(math.Max(float64(minNumSize), float64(pg.NumSize)))
            playing.Pokers = latestPokers()
            return
        }
    }

    if debug {
        tool.PrintVar(fmt.Sprintf("第%d次：玩家牌型非法", times), pokers)
    }
    err = errors.New("牌型不支持")
    return
}
