package job

import (
    "beginning/pkg/queue"
    "fmt"
    "github.com/golang-module/carbon"
)

type MyJobDemo struct {
    Time   carbon.Carbon `json:"time"`
    UserId int64         `json:"user_id"`
}

func (mjd *MyJobDemo) Handler() (queueErr *queue.Error) {
    defer func() {
        if err := recover(); err != nil {
            queueErr = queue.Err(fmt.Errorf("error for `MyJobDemo`: %+v", err))
        }
    }()

    // TODO 消费队列数据

    return
}
