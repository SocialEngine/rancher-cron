package metadata

import (
	"fmt"
	"time"

	"github.com/socialengine/rancher-cron/model"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	metadataURL = "http://rancher-metadata/latest"
)

// Client is a Struct that holds all metadata-specific data
type Client struct {
	MetadataClient  *metadata.Client
	EnvironmentName string
	CronLabelName   string
	Schedules       *model.Schedules
}

// NewClient creates a new metadata client
func NewClient(cronLabelName string) (*Client, error) {
	m, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		logrus.Fatalf("Failed to configure rancher-metadata: %v", err)
	}

	envName, err := getEnvironmentName(m)
	if err != nil {
		logrus.Fatalf("Error reading stack info: %v", err)
	}

	schedules := make(model.Schedules)

	return &Client{
		MetadataClient:  m,
		EnvironmentName: envName,
		CronLabelName:   cronLabelName,
		Schedules:       &schedules,
	}, nil
}

// GetVersion grabs the version of metadata client we're using
func (m *Client) GetVersion() (string, error) {
	return m.MetadataClient.GetVersion()
}

// GetCronSchedules returns a map of schedules with ContainerUUID as a key
func (m *Client) GetCronSchedules() (*model.Schedules, error) {
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
		containerUUIDs, err := m.getCronContainers(service)
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
			schedules[containerUUID] = schedule
		}
	}

	return &schedules, nil
}

func (m *Client) getCronContainers(service metadata.Service) ([]string, error) {
	containers := service.Containers
	var uuids []string

	for _, container := range containers {
		if len(container.ServiceName) == 0 {
			continue
		}

		if len(service.Name) != 0 {
			if service.Name != container.ServiceName {
				continue
			}
			if service.StackName != container.StackName {
				continue
			}
		}

		uuids = append(uuids, container.UUID)
	}

	if len(uuids) > 0 {
		return uuids, nil
	}

	return uuids, fmt.Errorf("could not find container UUID with %s", m.CronLabelName)
}

func markScheduleForCleanup(schedules *model.Schedules) {
	for _, schedule := range *schedules {
		schedule.ToCleanup = true
	}
}

func getEnvironmentName(m *metadata.Client) (string, error) {
	timeout := 30 * time.Second
	var err error
	var stack metadata.Stack
	for i := 1 * time.Second; i < timeout; i *= time.Duration(2) {
		stack, err = m.GetSelfStack()
		if err != nil {
			logrus.Errorf("Error reading stack info: %v...will retry", err)
			time.Sleep(i)
		} else {
			return stack.EnvironmentName, nil
		}
	}
	return "", fmt.Errorf("Error reading stack info: %v", err)
}
