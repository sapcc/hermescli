package audit

import (
	"net/http"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/client"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/sapcc/hermes-ctl/audit"
)

// configureDebug will configure the provider client to print the API
// requests and responses if OS_DEBUG is enabled.
func configureDebug(provider *gophercloud.ProviderClient) *gophercloud.ProviderClient {
	if os.Getenv("OS_DEBUG") != "" {
		provider.HTTPClient = http.Client{
			Transport: &client.RoundTripper{
				Rt:     &http.Transport{},
				Logger: &client.DefaultLogger{},
			},
		}
	}

	return provider
}

// NewHermesV1Client returns a *ServiceClient for making calls
// to the OpenStack Hermes v1 API. An error will be returned if
// authentication or client creation was not possible.
func NewHermesV1Client() (*gophercloud.ServiceClient, error) {
	ao, err := clientconfig.AuthOptions(nil)
	if err != nil {
		return nil, err
	}

	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	provider = configureDebug(provider)

	err = openstack.Authenticate(provider, *ao)
	if err != nil {
		return nil, err
	}

	return audit.NewHermesV1(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}
