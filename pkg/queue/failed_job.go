package queue

import "github.com/golang-module/carbon"

type FailedJobs struct {
	Id         int64         `gorm:"column:id" json:"id"`
	Connection string        `gorm:"column:connection;type:text"`
	Queue      string        `gorm:"column:queue;type:text"`
	Message    string        `gorm:"column:message;type:text"`
	Exception  string        `gorm:"column:exception;type:longText"`
	Stack      string        `gorm:"column:stack;type:longText"`
	FiledAt    carbon.Carbon `gorm:"column:failed_at"`
}

func (*FailedJobs) TableName() string {
	return "failed_jobs"
}
