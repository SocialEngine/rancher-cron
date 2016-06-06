package main

import (
	"flag"
	"os"

	"gopkg.in/robfig/cron.v2"

	"github.com/nowait-tools/rancher-cron/metadata"
	"github.com/socialengine/rancher-cron/cattle"
	"github.com/socialengine/rancher-cron/model"

	"github.com/Sirupsen/logrus"
)

const (
	poll = 1000
	// if metadata wasn't updated in 1 min, force update would be executed
	forceUpdateInterval = 10
	cronLabelName       = "com.socialengine.rancher-cron.schedule"
)

var (
	debug = flag.Bool("debug", false, "Debug")

	c  *cattle.Client
	m  *metadata.Client
	cr *cron.Cron
)

func setEnv() {
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	cattleURL := os.Getenv("CATTLE_URL")
	if len(cattleURL) == 0 {
		logrus.Fatalf("CATTLE_URL is not set")
	}

	cattleAPIKey := os.Getenv("CATTLE_ACCESS_KEY")
	if len(cattleAPIKey) == 0 {
		logrus.Fatalf("CATTLE_ACCESS_KEY is not set")
	}

	cattleSecretKey := os.Getenv("CATTLE_SECRET_KEY")
	if len(cattleSecretKey) == 0 {
		logrus.Fatalf("CATTLE_SECRET_KEY is not set")
	}

	//configure cattle client
	cClient, err := cattle.NewClient(cattleURL, cattleAPIKey, cattleSecretKey)

	if err != nil {
		logrus.Fatalf("Failed to configure cattle client: %v", err)
	}
	c = cClient

	// configure metadata client
	mClient, err := metadata.NewClient(cronLabelName)
	if err != nil {
		logrus.Fatalf("Failed to configure rancher-metadata client: %v", err)
	}
	m = mClient
}

func main() {
	logrus.Infof("Starting up Rancher Cron service")

	setEnv()

	go startHealthcheck()

	cr = cron.New()
	cr.AddFunc("*/30 * * * * *", discoverCronContainers)
	discoverCronContainers()
	cr.Start()

	select {}
}

func discoverCronContainers() {
	sched, err := m.GetCronSchedules()
	schedules := *sched

	if err != nil {
		logrus.Error("Could not retrieve cron schedules from rancher metadata service")
	}

	if len(schedules) == 0 {
		logrus.Errorf("Could not find any containers with label %s", cronLabelName)
	}

	// clean up old entities
	clearCron()

	logrus.Debugf("running discovery, found %d schedule containers", len(schedules))

	for _, schedule := range schedules {
		// we already have a cron job, and it was not cleaned up
		if schedule.CronID > 0 {
			continue
		}
		scheduleFunc, err := getCronFunction(schedule)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"container": schedule.ContainerUUID,
			}).Errorf("could not find container")
		}

		entryID, _ := cr.AddFunc(schedule.CronExpression, scheduleFunc)
		schedule.CronID = entryID
	}
}

func clearCron() {
	cleanedJobs := 0
	for key, schedule := range *m.Schedules {
		if !schedule.ToCleanup {
			continue
		}
		cr.Remove(schedule.CronID)
		delete(*m.Schedules, key)
		logrus.WithFields(logrus.Fields{
			"schedule":      schedule.CronExpression,
			"containerUUID": schedule.ContainerUUID,
		}).Info("deleted container from cron")

		cleanedJobs++
	}
	if cleanedJobs > 0 {
		logrus.Debugf("removed %d cron entries, and their schedule", cleanedJobs)
	}
}

func getCronFunction(schedule *model.Schedule) (func(), error) {
	container, err := c.GetContainerByUUID(schedule.ContainerUUID)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"schedule":      schedule.CronExpression,
		"containerName": container.Name,
		"containerUUID": schedule.ContainerUUID,
	}).Info("adding container to cron")

	return func() {
		container, err := c.StartContainerByID(container.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"containerId":   container.Id,
				"containerUuid": container.Uuid,
			}).Errorf("Failed to start container: %v", err)
		} else {
			logrus.WithFields(logrus.Fields{
				"containerId":   container.Id,
				"containerUuid": container.Uuid,
			}).Debugf("started container: %s", container.Name)
		}
	}, nil
}
