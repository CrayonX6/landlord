package variables

import (
    tool "beginning/pkg/acme"
    "fmt"
    "github.com/sony/sonyflake"
    "github.com/spf13/viper"
    rdsV4 "gopkg.in/redis.v4"
)

var SVariables Variables

type Variables struct {
}

var Snowflake *sonyflake.Sonyflake
var Rds *rdsV4.Client

func (v *Variables) Bootstrap() {
    Snowflake = sonyflake.NewSonyflake(sonyflake.Settings{})
    Rds = v.GetRds("redis")
}

// 获取 redis instance
func (v *Variables) GetRds(key string) *rdsV4.Client {
    cnf := viper.GetStringMapString(key)
    rds := rdsV4.NewClient(&rdsV4.Options{
        Addr:     fmt.Sprintf("%s:%s", cnf["host"], cnf["port"]),
        DB:       tool.ToInt(cnf["database"]),
        Password: cnf["auth"],
        PoolSize: 1000,
    })
    return rds
}
