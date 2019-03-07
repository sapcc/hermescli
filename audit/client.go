package audit

import (
	"github.com/gophercloud/gophercloud"
)

// NewHermesV1 creates a ServiceClient that may be used with the v1 hermes package.
func NewHermesV1(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	sc := new(gophercloud.ServiceClient)
	endpointOpts.ApplyDefaults("audit-data")
	url, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return sc, err
	}

	resourceBase := url // TODO: check the slash: + "/"
	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       url,
		Type:           "audit-data",
		ResourceBase:   resourceBase,
	}, nil
}
