// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package dataplaneconfig

import "github.com/gophercloud/gophercloud/v2"

func resourceURL(c *gophercloud.ServiceClient, projectID string) string {
	return c.ServiceURL("projects", projectID, "dataplane-config")
}
