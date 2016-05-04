package model

import (
	"github.com/rancher/go-rancher/client"
	"gopkg.in/robfig/cron.v2"
)

type Schedule struct {
	ToCleanup      bool
	CronExpression string
	ContainerUUID  string
	CronID         cron.EntryID
	Container      client.Container
}

type Schedules map[string]*Schedule

func NewSchedule() *Schedule {
	return &Schedule{
		ToCleanup: false,
	}
}
