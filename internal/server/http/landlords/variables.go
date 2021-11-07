package landlords

import (
    "gopkg.in/olahol/melody.v1"
)

// 单牌结构体
type Landlords struct {
    Image      string `json:"image"`       // 图片名
    Design     string `json:"design"`      // 花色
    Size       int    `json:"size"`        // 数值和花色的综合大小
    NumSize    int    `json:"num_size"`    // 数值大小
    DesignSize int    `json:"design_size"` // 花色大小
    Char       string `json:"char"`        // 字符标识
    Lord       bool   `json:"lord"`        // 该牌是否为地主牌
}

type PlayingPokers struct {
    Mode     string   `json:"mode"` // 出的牌型
    ModeSize int      `json:"mode_size"`
    NumSize  int      `json:"num_size"`
    Sound    []string `json:"sound"`
    Pokers   []string `json:"pokers"`
}

var ready = make(map[int64]bool)                    // 用户id 是否准备
var uid2rate = make(map[int64]int)                  // 用户设定的炸弹概率
var uid2rid = make(map[int64]int64)                 // 通过用户id -> 房间id
var uid2sid = make(map[int64]*melody.Session)       // 通过用户id -> 会话id
var sid2uid = make(map[*melody.Session]int64)       // 通过会话id -> 用户id
var rid int64 = 10001                               // 房间id
var rid2uid = make(map[int64][]int64)               // 通过房间id -> 用户id
var rid2surplus = make(map[int64]int64)             // 通过房间id -> 该房间剩余游戏人数
var uid2pokers = make(map[int64][]Landlords)        // 用户id -> 对局牌
var rid2lordPokers = make(map[int64][]Landlords)    // 房间id -> 底牌
var rid2lordUid = make(map[int64]int64)             // 房间id -> 地主用户id
var rid2grabNum = make(map[int64]int64)             // 房间id -> 第几个人叫地主
var rid2grabLog = make(map[int64][]bool)            // 房间id -> 每次叫地主的结果
var rid2grabToken = make(map[int64]map[string]bool) // 房间id -> 叫/抢地主TOKEN

var rid2playingNum = make(map[int64]int64)                   // 房间id -> 第几次出牌
var rid2playingLog = make(map[int64]map[int64]PlayingPokers) // 房间id -> 出牌日志
var rid2playingToken = make(map[int64]map[string]bool)       // 房间id -> 出牌TOKEN

// 玩家用户信息结构体
type UserInfo struct {
    Nickname  string                   `json:"nickname"`
    Uid       int64                    `json:"uid"`
    Avatar    string                   `json:"avatar"`
    Bean      int                      `json:"bean"`
    Lord      bool                     `json:"lord"`       // 该用户是否为地主
    Role      string                   `json:"role"`       // 角色描述信息
    PokersNum string                   `json:"pokers_num"` // 牌的数量描述
    Pokers    []map[string]interface{} `json:"pokers"`
    Skin      string                   `json:"skin"` // 音效皮肤
}

// 发牌三张可决定地主的情况
var map3 = map[string]int64{
    "000": 0,
    "001": 2,
    "010": 1,
    "100": 0,
}

// 发牌四张时地主的情况
var map4 = map[string]int64{
    "0110": 2,
    "0111": 1,
    "1010": 2,
    "1011": 0,
    "1100": 1,
    "1101": 0,
    "1110": 2,
    "1111": 0,
}

type PokerMode struct {
    Size  int      `json:"size"`  // 牌型大小
    Sound []string `json:"sound"` // 声音
    Same  []string `json:"same"`  // 指定某几个牌型为同类牌型
    DaNi  bool     `json:"da_ni"` // 使用"大你"音效
}

// 牌型
var pokerMode = map[string]PokerMode{
    "1":                      {Size: 1, Sound: []string{"sound/give"}},                                     // 单牌
    "11":                     {Size: 1, Sound: []string{"sound/give"}},                                     // 一对
    "111":                    {Size: 1, Sound: []string{"sound/give"}},                                     // 三个
    "111_1dan":               {Size: 1, Sound: []string{"sound/give", "sandaiyi"}, DaNi: true},             // 三带一单
    "111_1dui":               {Size: 1, Sound: []string{"sound/give", "sandaiyidui"}, DaNi: true},          // 三带一对
    "1111":                   {Size: 2, Sound: []string{"sound/give", "sound/boom", "zhadan"}},             // 炸弹
    "1111_2dan":              {Size: 1, Sound: []string{"sound/give", "sidaier"}, DaNi: true},              // 炸弹带两单
    "1111_2dui":              {Size: 1, Sound: []string{"sound/give", "sidailiangdui"}, DaNi: true},        // 炸弹带两对
    "11112222":               {Size: 1, Same: []string{"111222_2dan", "1111_2dui"}},                        // 特殊牌型
    "111122223333":           {Size: 1, Same: []string{"111222333_3dan"}},                                  // 特殊牌型
    "1111222233334444":       {Size: 1, Same: []string{"111222333444_4dan"}},                               // 特殊牌型
    "11112222333344445555":   {Size: 1, Same: []string{"111222333444555_5dan"}},                            // 特殊牌型
    "12345":                  {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "123456":                 {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "1234567":                {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "12345678":               {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "123456789":              {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "12345678910":            {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "1234567891011":          {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "123456789101112":        {Size: 1, Sound: []string{"sound/give", "shunzi"}, DaNi: true},               // 顺子
    "112233":                 {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "11223344":               {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "1122334455":             {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "112233445566":           {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "11223344556677":         {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "1122334455667788":       {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "112233445566778899":     {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "1122334455667788991010": {Size: 1, Sound: []string{"sound/give", "liandui"}, DaNi: true},              // 连对
    "111222":                 {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机不带
    "111222333":              {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机不带
    "111222333444":           {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机不带
    "111222333444555":        {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机不带
    "111222333444555666":     {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机不带
    "111222_2dan":            {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带单
    "111222333_3dan":         {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带单
    "111222333444_4dan":      {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带单
    "111222333444555_5dan":   {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带单
    "111222_2dui":            {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带对
    "111222333_3dui":         {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带对
    "111222333444_4dui":      {Size: 1, Sound: []string{"sound/give", "sound/plane", "feiji"}, DaNi: true}, // 飞机带对
    "xy":                     {Size: 3, Sound: []string{"sound/give", "sound/boom", "wangzha"}},            // 王炸
}
