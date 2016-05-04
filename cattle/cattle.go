package cattle

import (
	"fmt"
	"github.com/rancher/go-rancher/client"
)

type CattleClient struct {
	rancherClient *client.RancherClient
}

func NewCattleClient(cattleUrl string, cattleAccessKey string, cattleSecretKey string) (*CattleClient, error) {
	apiClient, err := client.NewRancherClient(&client.ClientOpts{
		Url:       cattleUrl,
		AccessKey: cattleAccessKey,
		SecretKey: cattleSecretKey,
	})

	if err != nil {
		return nil, err
	}

	return &CattleClient{
		rancherClient: apiClient,
	}, nil
}

func (c *CattleClient) GetContainerByUUID(uuid string) (*client.Container, error) {
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

func (c *CattleClient) StartContainer(container *client.Container) (*client.Instance, error) {
	if container.State != "stopped" {
		return nil, fmt.Errorf("container is not stopped. Currently in [%s] state", container.State)
	}
	return c.rancherClient.Container.ActionStart(container)
}

func (c *CattleClient) TestConnect() error {
	opts := &client.ListOpts{}
	_, err := c.rancherClient.Label.List(opts)
	return err
}
