package metadata

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	metadataURL = "http://rancher-metadata/latest"
)

// Client is a Struct that holds all metadata-specific data
type Client struct {
	MetadataClient  metadata.Client
	EnvironmentName string
	CronLabelName   string
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

	return &Client{
		MetadataClient:  m,
		EnvironmentName: envName,
		CronLabelName:   cronLabelName,
	}, nil
}

// GetVersion grabs the version of metadata client we're using
func (m *Client) GetVersion() (string, error) {
	return m.MetadataClient.GetVersion()
}

// GetServices returns the services from the metadata client
func (m *Client) GetServices() ([]metadata.Service, error) {
	return m.MetadataClient.GetServices()
}

// GetContainersFromService returns an array of UUIDs for the containers within a service
func (m *Client) GetContainersFromService(service metadata.Service) ([]string, error) {
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

func getEnvironmentName(m metadata.Client) (string, error) {
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
