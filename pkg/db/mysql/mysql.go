package mysql

import (
    "beginning/pkg/db"
    "fmt"
    "time"

    driver_mysql "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/plugin/prometheus"
)

type mysql struct {
    connection string
    database   string
    host       string
    port       string
    username   string
    password   string
    debug      bool
}

func NewMysql(connection, database, host, port, username, password string, debug bool) *mysql {
    return &mysql{connection, database, host, port, username, password, debug}
}

func (m *mysql) Connect() {
    conn := m.connection
    database := m.database
    host := m.host
    port := m.port
    user := m.username
    pwd := m.password
    debug := m.debug

    dsn := user + ":" + pwd + "@(" + host + ":" + port + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
    for {
        loglevel := logger.Info
        if !debug {
            loglevel = logger.Silent
        }
        orm, err := gorm.Open(driver_mysql.Open(dsn), &gorm.Config{
            Logger:                                   logger.Default.LogMode(loglevel),
            SkipDefaultTransaction:                   true,
            PrepareStmt:                              true,
            DisableForeignKeyConstraintWhenMigrating: true,
        })
        if err == nil {
            _ = orm.Use(prometheus.New(prometheus.Config{
                DBName:          "beginning",
                RefreshInterval: 15,
                StartServer:     false,
                HTTPServerPort:  9002,
                MetricsCollector: []prometheus.MetricsCollector{
                    &prometheus.MySQL{
                        Prefix:        "gorm_status_",
                        Interval:      100,
                        VariableNames: []string{"Threads_running"},
                    },
                },
            }))

            sqlDB, _ := orm.DB()
            //default conn
            sqlDB.SetMaxIdleConns(10)
            sqlDB.SetMaxOpenConns(1000)
            sqlDB.SetConnMaxLifetime(time.Hour)
            db.Conns.Store(conn, orm)
            break
        }
        fmt.Printf("Database connection exception! 5 seconds to retry, errors = %v \r\n", err)
        time.Sleep(5 * time.Second)
    }
}
