// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

// Package gophercloud-sapcc provides integration between SAP CC services and
// Gophercloud.
package clients

import (
	"github.com/gophercloud/gophercloud/v2"
)

// NewLimesV1 creates a ServiceClient that may be used to interact with Limes' API
// endpoints that deal with resources.
func NewLimesV1(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	endpointOpts.ApplyDefaults("resources")
	endpoint, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return nil, err
	}

	endpoint += "v1/"

	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       endpoint,
		Type:           "resources",
	}, nil
}

// NewLimesRatesV1 creates a ServiceClient that may be used to interact with Limes' API
// endpoints that deal with rates.
func NewLimesRatesV1(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	endpointOpts.ApplyDefaults("sapcc-rates")
	endpoint, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return nil, err
	}
	endpoint += "v1/"

	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       endpoint,
		Type:           "sapcc-rates",
	}, nil
}

// NewAutomationV1 creates a ServiceClient that may be used with the v1 automation package.
func NewAutomationV1(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	sc := new(gophercloud.ServiceClient)
	endpointOpts.ApplyDefaults("automation")
	url, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return sc, err
	}

	resourceBase := url + "api/v1/"
	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       url,
		Type:           "automation",
		ResourceBase:   resourceBase,
	}, nil
}

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

// NewBilling creates a ServiceClient that may be used with the billing package.
func NewBilling(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	sc := new(gophercloud.ServiceClient)
	endpointOpts.ApplyDefaults("sapcc-billing")
	url, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return sc, err
	}

	resourceBase := url
	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       url,
		Type:           "sapcc-billing",
		ResourceBase:   resourceBase,
	}, nil
}

// NewArcV1 creates a ServiceClient that may be used with the v1 arc package.
func NewArcV1(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	sc := new(gophercloud.ServiceClient)
	endpointOpts.ApplyDefaults("arc")
	url, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return sc, err
	}

	resourceBase := url + "api/v1/"
	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       url,
		Type:           "arc",
		ResourceBase:   resourceBase,
	}, nil
}

// NewMetisV1 creates a ServiceClient that may be used with the v1 metis package.
func NewMetisV1(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	sc := new(gophercloud.ServiceClient)
	endpointOpts.ApplyDefaults("metis")
	url, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return sc, err
	}

	resourceBase := url + "v1/"
	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       url,
		Type:           "metis",
		ResourceBase:   resourceBase,
	}, nil
}
