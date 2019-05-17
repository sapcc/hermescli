package audit

import (
	"net/http"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/sapcc/hermes-ctl/audit"
	"github.com/sapcc/hermes-ctl/client"
)

// configureDebug will configure the provider client to print the API
// requests and responses if OS_DEBUG is enabled.
func configureDebug(client *gophercloud.ProviderClient) *gophercloud.ProviderClient {
	if os.Getenv("OS_DEBUG") != "" {
		client.HTTPClient = http.Client{
			Transport: &clients.LogRoundTripper{
				Rt: &http.Transport{},
			},
		}
	}

	return client
}

// NewHermesV1Client returns a *ServiceClient for making calls
// to the OpenStack Hermes v1 API. An error will be returned if
// authentication or client creation was not possible.
func NewHermesV1Client() (*gophercloud.ServiceClient, error) {
	ao, err := client.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	client = configureDebug(client)

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return nil, err
	}

	return audit.NewHermesV1(client, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}
