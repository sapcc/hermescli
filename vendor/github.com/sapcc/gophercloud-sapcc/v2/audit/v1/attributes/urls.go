// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package attributes

import "github.com/gophercloud/gophercloud/v2"

func listURL(c *gophercloud.ServiceClient, name string) string {
	return c.ServiceURL("attributes", name)
}
