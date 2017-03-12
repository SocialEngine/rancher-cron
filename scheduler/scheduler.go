package scheduler

import (
	"github.com/Sirupsen/logrus"

	"github.com/socialengine/rancher-cron/cattle"
	"github.com/socialengine/rancher-cron/metadata"
	"github.com/socialengine/rancher-cron/model"
)

// Scheduler is a Struct and contains all the schedule info
type Scheduler struct {
	CattleClient   *cattle.Client
	MetadataClient *metadata.Client
	CronLabelName  string
	Schedules      *model.Schedules
}

// NewScheduler creates a new Scheduler
func NewScheduler(cronLabelName string, metadataClient *metadata.Client, cattleClient *cattle.Client) (*Scheduler, error) {
	schedules := make(model.Schedules)

	return &Scheduler{
		CattleClient:   cattleClient,
		MetadataClient: metadataClient,
		CronLabelName:  cronLabelName,
		Schedules:      &schedules,
	}, nil
}

// GetCronSchedules returns a map of schedules with ContainerUUID as a key
func (m *Scheduler) GetCronSchedules() (*model.Schedules, error) {
	schedules := *m.Schedules
	services, err := m.MetadataClient.GetServices()

	if err != nil {
		logrus.Infof("Error reading services %v", err)
		return &schedules, err
	}

	markScheduleForCleanup(m.Schedules)

	for _, service := range services {
		cronExpression, ok := service.Labels[m.CronLabelName]
		if !ok {
			continue
		}

		cattleService, err := m.CattleClient.GetServiceByUUID(service.UUID)
		if err == nil && cattleService.State == "inactive" {
			logrus.Debugf("ignoring inactive service uuid %s", service.UUID)
			continue
		}

		containerUUIDs, err := m.MetadataClient.GetContainersFromService(service)
		if err != nil {
			continue
		}

		for _, containerUUID := range containerUUIDs {
			existingSchedule, ok := schedules[containerUUID]
			// we already have schedule for this container
			if ok {
				// do not cleanup
				existingSchedule.ToCleanup = false

				logrus.WithFields(logrus.Fields{
					"uuid":           containerUUID,
					"cronExpression": cronExpression,
				}).Debugf("already have container")

				continue
			}
			//label exists, configure schedule
			schedule := model.NewSchedule()

			schedule.CronExpression = cronExpression
			schedule.ContainerUUID = containerUUID
			schedule.ServiceUUID = service.UUID
			schedules[containerUUID] = schedule
		}
	}

	return &schedules, nil
}

func markScheduleForCleanup(schedules *model.Schedules) {
	for _, schedule := range *schedules {
		schedule.ToCleanup = true
	}
}
