package entity

import (
    "fmt"
    "github.com/golang-module/carbon"
)

type MySQLTable struct {
    ID         int64         `json:"id"`
    AddTime    carbon.Carbon `gorm:"autoCreateTime;" json:"-"`
    UpdateTime carbon.Carbon `gorm:"autoUpdateTime;" json:"-"`
    State      int8          `gorm:"default:1;" json:"-"`
}

// Cache key for 用户实时行为 WebSocket
func CK4UserRealTimeBehavior(userId int64) string {
    return fmt.Sprintf("cw:user_real_time_behavior:%d", userId)
}

// Cache key for 全服实时行为 WebSocket
func CK4AllRealTimeBehavior() string {
    return "cw:all_real_time_behavior"
}
