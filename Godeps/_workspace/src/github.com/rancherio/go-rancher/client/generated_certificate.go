package client

const (
	CERTIFICATE_TYPE = "certificate"
)

type Certificate struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Cert string `json:"cert,omitempty" yaml:"cert,omitempty"`

	CertChain string `json:"certChain,omitempty" yaml:"cert_chain,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Key string `json:"key,omitempty" yaml:"key,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type CertificateCollection struct {
	Collection
	Data []Certificate `json:"data,omitempty"`
}

type CertificateClient struct {
	rancherClient *RancherClient
}

type CertificateOperations interface {
	List(opts *ListOpts) (*CertificateCollection, error)
	Create(opts *Certificate) (*Certificate, error)
	Update(existing *Certificate, updates interface{}) (*Certificate, error)
	ById(id string) (*Certificate, error)
	Delete(container *Certificate) error
}

func newCertificateClient(rancherClient *RancherClient) *CertificateClient {
	return &CertificateClient{
		rancherClient: rancherClient,
	}
}

func (c *CertificateClient) Create(container *Certificate) (*Certificate, error) {
	resp := &Certificate{}
	err := c.rancherClient.doCreate(CERTIFICATE_TYPE, container, resp)
	return resp, err
}

func (c *CertificateClient) Update(existing *Certificate, updates interface{}) (*Certificate, error) {
	resp := &Certificate{}
	err := c.rancherClient.doUpdate(CERTIFICATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *CertificateClient) List(opts *ListOpts) (*CertificateCollection, error) {
	resp := &CertificateCollection{}
	err := c.rancherClient.doList(CERTIFICATE_TYPE, opts, resp)
	return resp, err
}

func (c *CertificateClient) ById(id string) (*Certificate, error) {
	resp := &Certificate{}
	err := c.rancherClient.doById(CERTIFICATE_TYPE, id, resp)
	return resp, err
}

func (c *CertificateClient) Delete(container *Certificate) error {
	return c.rancherClient.doResourceDelete(CERTIFICATE_TYPE, &container.Resource)
}
