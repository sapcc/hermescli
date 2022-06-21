// Copyright 2019 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package attributes

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToAttributeListQuery() (string, error)
}

// ListOpts allows the filtering of paginated collections through the API.
// Filtering is achieved by passing in filter value. Page and PerPage are used
// for pagination.
type ListOpts struct {
	MaxDepth  int    `q:"max_depth"`
	Limit     int    `q:"limit"`
	DomainID  string `q:"domain_id"`
	ProjectID string `q:"project_id"`
}

// ToAttributeListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToAttributeListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// List retrieves a list of Attributes.
func List(client *gophercloud.ServiceClient, name string, opts ListOptsBuilder) pagination.Pager {
	url := listURL(client, name)
	if opts != nil {
		query, err := opts.ToAttributeListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return AttributePage{pagination.SinglePageBase(r)}
	})
}
