package cattle

import (
	"fmt"

	"github.com/rancher/go-rancher/client"
)

// Client holds rancherClient and anything else that could be useful
type Client struct {
	rancherClient *client.RancherClient
}

// NewClient grabs config necessary and sets an inited client or returns an error
func NewClient(cattleURL string, cattleAccessKey string, cattleSecretKey string) (*Client, error) {
	apiClient, err := client.NewRancherClient(&client.ClientOpts{
		Url:       cattleURL,
		AccessKey: cattleAccessKey,
		SecretKey: cattleSecretKey,
	})

	if err != nil {
		return nil, err
	}

	return &Client{
		rancherClient: apiClient,
	}, nil
}

// GetServiceByUUID grabs a Service given a UUID
func (c *Client) GetServiceByUUID(uuid string) (*client.Service, error) {
	opts := &client.ListOpts{
		Filters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	services, err := c.rancherClient.Service.List(opts)

	if err != nil {
		return nil, err
	}

	service := &services.Data[0]

	return service, err
}

// GetContainerByUUID grabs a Container given a UUID
func (c *Client) GetContainerByUUID(uuid string) (*client.Container, error) {
	opts := &client.ListOpts{
		Filters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	containers, err := c.rancherClient.Container.List(opts)

	if err != nil {
		return nil, err
	}

	container := &containers.Data[0]

	return container, err
}

// StartContainerByID starts a container given ContainerId or fails in cases its already running
func (c *Client) StartContainerByID(containerID string) (*client.Container, error) {
	container, err := c.rancherClient.Container.ById(containerID)

	if err != nil {
		return container, err
	}

	if container.State == "stopped" {
		_, err = c.rancherClient.Container.ActionStart(container)
	} else if container.State == "running" {
		_, err = c.rancherClient.Container.ActionRestart(container)
	} else {
		err = fmt.Errorf("container is not stopped and running. Currently in [%s] state", container.State)
	}

	return container, err
}

// TestConnect ensures we can query the API
func (c *Client) TestConnect() error {
	opts := &client.ListOpts{}
	_, err := c.rancherClient.Label.List(opts)
	return err
}
