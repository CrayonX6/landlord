package acme

import (
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "fmt"
    "github.com/golang-module/carbon"
    rdsV4 "gopkg.in/redis.v4"
    "time"
)

// sort-set 写入一个排行数据
func (a *Acme) SSRankPush(rankKey string, score float64, member interface{}) {
    variables.Rds.ZAdd(rankKey, rdsV4.Z{Score: score, Member: member})
}

// sort-set 移除一个成员
func (a *Acme) SSRankRemove(rankKey string, member interface{}) {
    variables.Rds.ZRem(rankKey, member)
}

// sort-set 获取成员分数
func (a *Acme) SSRankScore(rankKey string, member interface{}) float64 {
    return variables.Rds.ZScore(rankKey, tool.ToStr(member)).Val()
}

// sort-set 获取指定标识的排行名次
func (a *Acme) SSRankIndex(rankKey string, member interface{}) int64 {
    cmd := variables.Rds.ZRank(rankKey, tool.ToStr(member))
    if cmd.Err() != nil {
        return 0
    }
    return cmd.Val() + 1
}

// sort-set 获取指定范围的排行列表
func (a *Acme) SSRankList(rankKey string, page int64, pageSize int64) []string {
    if page < 1 {
        page = 1
    }
    start := (page - 1) * pageSize
    cmd := variables.Rds.ZRange(rankKey, start, start+pageSize-1)
    if cmd.Err() != nil {
        return []string{}
    }
    return cmd.Val()
}

// sort-set 获取指定范围的排行列表及分数
func (a *Acme) SSRankListWithScore(rankKey string, page int64, pageSize int64) []rdsV4.Z {
    if page < 1 {
        page = 1
    }
    start := (page - 1) * pageSize
    cmd := variables.Rds.ZRangeWithScores(rankKey, start, start+pageSize-1)
    if cmd.Err() != nil {
        return []rdsV4.Z{}
    }
    return cmd.Val()
}

// Redis 获取同步重试锁
// SetNX
func (a *Acme) GetSyncTryAgainLock(key string, lockTryTimes int, lockTryGapMs int) bool {
    if lockTryTimes <= 0 {
        return false
    }
    val := fmt.Sprintf("first lock at %s", carbon.Now().ToDateTimeString())
    lock := variables.Rds.SetNX(key, val, 0).Val()
    if !lock {
        time.Sleep(time.Millisecond * time.Duration(lockTryGapMs))
        return a.GetSyncTryAgainLock(key, lockTryTimes-1, lockTryGapMs)
    }
    return lock
}

// Redis 释放同步重试锁
func (a *Acme) UnSyncTryAgainLock(key string) {
    variables.Rds.Del(key)
}

// Redis 获取同步阻塞锁
// SetNX + List.BRPOP
func (a *Acme) GetSyncLock(key string, timeout time.Duration) bool {
    lock := variables.Rds.SetNX(key, fmt.Sprintf("first lock at %s", carbon.Now().ToDateTimeString()), 0).Val()
    if !lock {
        lKey := fmt.Sprintf("%s:sync", key)
        listLock := variables.Rds.BRPop(timeout, lKey).Val()
        if len(listLock) > 0 {
            lock = true
        }
    }
    return lock
}

// Redis 释放同步阻塞锁
// List.LPUSH
// Warning: 注意业务挂断出现死锁的情况，应该在 defer 中释放锁
func (a *Acme) UnSyncLock(key string) {
    lKey := fmt.Sprintf("%s:sync", key)
    variables.Rds.LPush(lKey, fmt.Sprintf("unlocked at %s", carbon.Now().ToDateTimeString()))
}
