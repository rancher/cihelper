package client

const (
	BASE_MACHINE_CONFIG_TYPE = "baseMachineConfig"
)

type BaseMachineConfig struct {
	Resource `yaml:"-"`
}

type BaseMachineConfigCollection struct {
	Collection
	Data   []BaseMachineConfig `json:"data,omitempty"`
	client *BaseMachineConfigClient
}

type BaseMachineConfigClient struct {
	rancherClient *RancherClient
}

type BaseMachineConfigOperations interface {
	List(opts *ListOpts) (*BaseMachineConfigCollection, error)
	Create(opts *BaseMachineConfig) (*BaseMachineConfig, error)
	Update(existing *BaseMachineConfig, updates interface{}) (*BaseMachineConfig, error)
	ById(id string) (*BaseMachineConfig, error)
	Delete(container *BaseMachineConfig) error
}

func newBaseMachineConfigClient(rancherClient *RancherClient) *BaseMachineConfigClient {
	return &BaseMachineConfigClient{
		rancherClient: rancherClient,
	}
}

func (c *BaseMachineConfigClient) Create(container *BaseMachineConfig) (*BaseMachineConfig, error) {
	resp := &BaseMachineConfig{}
	err := c.rancherClient.doCreate(BASE_MACHINE_CONFIG_TYPE, container, resp)
	return resp, err
}

func (c *BaseMachineConfigClient) Update(existing *BaseMachineConfig, updates interface{}) (*BaseMachineConfig, error) {
	resp := &BaseMachineConfig{}
	err := c.rancherClient.doUpdate(BASE_MACHINE_CONFIG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *BaseMachineConfigClient) List(opts *ListOpts) (*BaseMachineConfigCollection, error) {
	resp := &BaseMachineConfigCollection{}
	err := c.rancherClient.doList(BASE_MACHINE_CONFIG_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *BaseMachineConfigCollection) Next() (*BaseMachineConfigCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &BaseMachineConfigCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *BaseMachineConfigClient) ById(id string) (*BaseMachineConfig, error) {
	resp := &BaseMachineConfig{}
	err := c.rancherClient.doById(BASE_MACHINE_CONFIG_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *BaseMachineConfigClient) Delete(container *BaseMachineConfig) error {
	return c.rancherClient.doResourceDelete(BASE_MACHINE_CONFIG_TYPE, &container.Resource)
}
