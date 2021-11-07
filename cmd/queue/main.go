package main

import (
    "beginning/configs"
    "beginning/internal/server"
    "beginning/pkg/cache"
    "beginning/pkg/cache/redis"
    "beginning/pkg/cache/sync"
    "beginning/pkg/db"
    "beginning/pkg/db/mysql"
    "beginning/pkg/file/cos"
    "beginning/pkg/log/zap"
    "beginning/pkg/queue"
    "beginning/pkg/queue/rabbitmq"
    "fmt"
    "github.com/fsnotify/fsnotify"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/spf13/viper"
    _ "go.uber.org/automaxprocs"
    "io/ioutil"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"
)

func main() {
    initConfig()
    initLog()
    initFs()
    initDB()
    initCache()
    initQueue()
    initProvider()
    initMetrics()
    initSignal()
}

func initConfig() {
    env, _ := ioutil.ReadFile(".env")
    filename := fmt.Sprintf("config.%s.toml", strings.Trim(string(env), "\r\n "))
    _, err := os.Stat(filename)
    if (err == nil) || (os.IsExist(err)) {
        viper.SetConfigFile(filename)
        err := viper.ReadInConfig()
        if err != nil {
            panic(err)
        }
        viper.WatchConfig()
        viper.OnConfigChange(func(in fsnotify.Event) {
            configs.LoadConf()
        })
    }
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    configs.LoadConf()
}

func initLog() {
    env := viper.GetString("app.env")
    filepath := "./storage/log/"
    filename := fmt.Sprintf("api.%s.log", env)
    _, err := os.Stat(filepath + filename)
    if err != nil {
        if os.IsNotExist(err) {
            err := os.MkdirAll(filepath, os.ModePerm)
            if err != nil {
                panic(err)
            }
            f, err := os.Create(filepath + filename)
            if err != nil {
                panic(err)
            }
            _ = f.Close()
        }
    }
    zap.NewZaps([]string{filepath + filename, "stdout"}).InitLog()
}

func initFs() {
    fs := viper.Get("filesystem").(map[string]map[string]string)
    for _, v := range fs {
        switch v["driver"] {
        case "cos":
            co := cos.NewCosFile(v["secret_id"], v["secret_key"], v["region"], v["bucket"], 100*time.Second)
            co.InitCos()
        }
    }
}

func initDB() {
    con := viper.Get("db.connections").(map[string]map[string]string)
    defaultCon := viper.GetString("db.connection")
    debug := viper.GetBool("app.debug")
    for key, v := range con {
        switch v["driver"] {
        case "mysql":
            ms := mysql.NewMysql(key, v["database"], v["host"], v["port"], v["username"], v["password"], debug)
            ms.Connect()
        }
        if defaultCon == key {
            db.SetDefault(db.ORM(key))
        }
    }
}

func initCache() {
    caches := viper.Get("caches").(map[string]map[string]string)
    defaultCache := viper.GetString("cache.drive")
    for key, v := range caches {
        switch v["driver"] {
        case "redis":
            database, _ := strconv.Atoi(v["database"])
            c := redis.New(v["host"], v["port"], database, v["auth"])
            cache.CacheMap.Store(key, c)
        case "sync":
            c := sync.New()
            cache.CacheMap.Store(key, c)
        }
        if defaultCache == key {
            cache.SetDefault(cache.GetCache(key))
        }
    }
}

func initQueue() {
    caches := viper.Get("queues").(map[string]map[string]string)
    defaultQueue := viper.GetString("queue.drive")
    prefix := viper.GetString("app.name")
    for key, v := range caches {
        switch v["driver"] {
        case "rabbit-mq":
            r := rabbitmq.NewRabbitMQ(v["user"], v["pass"], v["host"], v["port"], v["vhost"], prefix)
            queue.QueueMap.Store(key, r)
        }
        if defaultQueue == key {
            queue.SetDefault(queue.GetMQ(key))
            queue.MQ.Connect()
        }
    }
}

func initProvider() {
    server.RepoSP.Register()
    server.JobSP.Register()
}

func initMetrics() {
    go func() {
        svr := http.NewServeMux()
        svr.Handle("/metrics", promhttp.Handler())
        _ = http.ListenAndServe(":9002", svr)
    }()
}

func initSignal() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
    for {
        s := <-c
        switch s {
        case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
            fmt.Print("api exit")
            return
        case syscall.SIGHUP:
        default:
            return
        }
    }
}
