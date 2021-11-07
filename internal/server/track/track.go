package track

import (
    "beginning/internal/server/job"
    "beginning/pkg/queue"
    "context"
    "github.com/golang-module/carbon"
    "time"
)

var STrack Track

type Track struct {
}

// MyJobDemo
func (t *Track) Track4MyJobDemo(userId int64) {
    ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
    go func() {
        queue.NewProducer("my_job_demo", &job.MyJobDemo{
            Time:   carbon.Now(),
            UserId: userId,
        })
        ctx.Done()
    }()
}
