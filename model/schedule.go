package model

import "gopkg.in/robfig/cron.v2"

// Schedule holds data related to an individual schedule such as cronId, schedule, ContainerUUID, etc
type Schedule struct {
	ToCleanup      bool
	CronExpression string
	ContainerUUID  string
	CronID         cron.EntryID
	ServiceUUID    string
}

// Schedules is a simple collection of Schedule
type Schedules map[string]*Schedule

// NewSchedule is a constructor to help set default values
func NewSchedule() *Schedule {
	return &Schedule{
		ToCleanup: false,
	}
}
