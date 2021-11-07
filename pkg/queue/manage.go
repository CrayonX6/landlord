package queue

import (
	"beginning/pkg/db"
	"sync"
)

var QueueMap sync.Map

var MQ Queue

var ErrJob chan FailedJobs

func SetDefault(g Queue) {
	MQ = g
}

func GetMQ(c string) Queue {
	v, ok := QueueMap.Load(c);
	if ok {
		return v.(Queue)
	}
	return nil
}

func NewConsumer(topic string, job JobBase, sleep, retry int32) {
	AutoMigrate()
	mq := MQ.Connect()
	mq.Topic(topic)
	mq.Consumer(job, sleep, retry)
}

func NewProducer(topic string, job JobBase) {
	mq := MQ.GetInstanceConnect()
	mq.Topic(topic)
	mq.Producer(job)
}

func AutoMigrate() {
	_ = db.Orm.AutoMigrate(&FailedJobs{})
	ErrJob = make(chan FailedJobs)
	go func() {
		for {
			select {
			case failedJob := <-ErrJob:
				db.Orm.Save(&failedJob)
			}
		}
	}()
}
