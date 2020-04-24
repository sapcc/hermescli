package events

import "github.com/gophercloud/gophercloud"

func eventURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("events", id)
}

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("events")
}

func listURL(c *gophercloud.ServiceClient) string {
	return rootURL(c)
}

func getURL(c *gophercloud.ServiceClient, id string) string {
	return eventURL(c, id)
}
