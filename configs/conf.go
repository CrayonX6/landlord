package configs

import (
    "github.com/spf13/viper"
)

func LoadConf() {
    viper.Set("db.connections", map[string]map[string]string{
        "mysql": {
            "driver":   "mysql",
            "host":     viper.GetString("db.host"),
            "port":     viper.GetString("db.port"),
            "database": viper.GetString("db.database"),
            "username": viper.GetString("db.username"),
            "password": viper.GetString("db.password"),
        },
    })

    viper.Set("filesystem", map[string]map[string]string{
        "cos": {
            "driver":     "cos",
            "bucket":     viper.GetString("fs.bucket"),
            "region":     viper.GetString("fs.region"),
            "secret_id":  viper.GetString("fs.secret_id"),
            "secret_key": viper.GetString("fs.secret_key"),
        },
    })

    viper.Set("caches", map[string]map[string]string{
        "redis": {
            "driver":   "redis",
            "host":     viper.GetString("redis.host"),
            "port":     viper.GetString("redis.port"),
            "database": viper.GetString("redis.database"),
            "auth":     viper.GetString("redis.auth"),
        },
        "sync": {
            "driver": "sync",
        },
    })

    viper.Set("queues", map[string]map[string]string{
        "rabbit-mq": {
            "driver": "rabbit-mq",
            "host":   viper.GetString("rabbit-mq.host"),
            "port":   viper.GetString("rabbit-mq.port"),
            "vhost":  viper.GetString("rabbit-mq.vhost"),
            "user":   viper.GetString("rabbit-mq.user"),
            "pass":   viper.GetString("rabbit-mq.pass"),
        },
    })
}
