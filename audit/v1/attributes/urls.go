package attributes

import "github.com/gophercloud/gophercloud"

func listURL(c *gophercloud.ServiceClient, name string) string {
	return c.ServiceURL("attributes", name)
}
