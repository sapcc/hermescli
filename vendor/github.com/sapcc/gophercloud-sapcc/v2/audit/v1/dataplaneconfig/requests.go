// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package dataplaneconfig

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
)

// PutOpts specifies the fields for a Put request.
type PutOpts struct {
	Enabled      bool   `json:"enabled"`
	TargetBucket string `json:"target_bucket,omitempty"`
}

// Get retrieves the dataplane configuration for the given project.
// Returns the default (disabled) config when none has been set.
func Get(ctx context.Context, c *gophercloud.ServiceClient, projectID string) (r GetResult) {
	//nolint:bodyclose // already handled by gophercloud
	resp, err := c.Get(ctx, resourceURL(c, projectID), &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{http.StatusOK},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Put creates or replaces the dataplane configuration for the given project.
func Put(ctx context.Context, c *gophercloud.ServiceClient, projectID string, opts PutOpts) (r PutResult) {
	//nolint:bodyclose // already handled by gophercloud
	resp, err := c.Put(ctx, resourceURL(c, projectID), opts, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{http.StatusOK},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Delete removes the dataplane configuration for the given project.
// Deleting a non-existent config is a no-op (returns 204).
func Delete(ctx context.Context, c *gophercloud.ServiceClient, projectID string) (r DeleteResult) {
	//nolint:bodyclose // already handled by gophercloud
	resp, err := c.Delete(ctx, resourceURL(c, projectID), &gophercloud.RequestOpts{
		OkCodes: []int{http.StatusNoContent},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
