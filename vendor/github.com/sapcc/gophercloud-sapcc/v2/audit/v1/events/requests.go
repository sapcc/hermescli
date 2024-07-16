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

package events

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// DateFilter represents a valid filter to use for filtering
// events by their date during a list.
type DateFilter string

const (
	DateFilterGT  DateFilter = "gt"
	DateFilterGTE DateFilter = "gte"
	DateFilterLT  DateFilter = "lt"
	DateFilterLTE DateFilter = "lte"
)

// DateQuery represents a date field to be used for listing events.
// If no filter is specified, the query will act as if "equal" is used.
type DateQuery struct {
	Date   time.Time
	Filter DateFilter
}

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToEventListQuery() (string, error)
}

// ListOpts allows the filtering of paginated collections through the API.
// Filtering is achieved by passing in filter value. Page and PerPage are used
// for pagination.
type ListOpts struct {
	ObserverType string `q:"observer_type"`
	TargetID     string `q:"target_id"`
	TargetType   string `q:"target_type"`
	InitiatorID  string `q:"initiator_id"`

	// Not available for sort
	InitiatorType string `q:"initiator_type"`
	InitiatorName string `q:"initiator_name"`

	Action      string `q:"action"`
	Outcome     string `q:"outcome"`
	RequestPath string `q:"request_path"`
	Time        []DateQuery
	DomainID    string `q:"domain_id"`
	ProjectID   string `q:"project_id"`

	// Search is over the entire event body.
	Search string `q:"search"`

	// Sort will sort the results in the requested order.
	Sort string `q:"sort"`

	Limit  int `q:"limit"`
	Offset int `q:"offset"`
}

// ToEventListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToEventListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	params := q.Query()

	if len(opts.Time) > 0 {
		var t []string

		for _, dt := range opts.Time {
			tmp := dt.Date.Format(time.RFC3339)
			if dt.Filter == "" {
				// combine gte and lte, when there is no filter
				t = append(t, fmt.Sprintf("%s:%s,%s:%s", DateFilterGTE, tmp, DateFilterLTE, tmp))
			} else {
				t = append(t, fmt.Sprintf("%s:%s", dt.Filter, tmp))
			}
		}

		params.Add("time", strings.Join(t, ","))
	}

	q = &url.URL{RawQuery: params.Encode()}

	return q.String(), nil
}

// List retrieves a list of Events.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	serviceURL := listURL(client)
	if opts != nil {
		query, err := opts.ToEventListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		serviceURL += query
	}
	return pagination.NewPager(client, serviceURL, func(r pagination.PageResult) pagination.Page {
		return EventPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// GetOptsBuilder allows extensions to add additional parameters to the
// Get request.
type GetOptsBuilder interface {
	ToEventsQuery() (string, error)
}

// GetOpts enables retrieving events by a specific project or domain.
type GetOpts struct {
	// The project ID to retrieve event for.
	ProjectID string `q:"project_id"`

	// The domain ID to retrieve event for.
	DomainID string `q:"domain_id"`
}

// ToEventsQuery formats a GetOpts into a query string.
func (opts GetOpts) ToEventsQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// Get retrieves a specific event based on its unique ID.
func Get(ctx context.Context, c *gophercloud.ServiceClient, id string, opts GetOptsBuilder) (r GetResult) {
	serviceURL := getURL(c, id)
	if opts != nil {
		query, err := opts.ToEventsQuery()
		if err != nil {
			r.Err = err
			return
		}
		serviceURL += query
	}

	//nolint:bodyclose // already handled by gophercloud
	resp, err := c.Get(ctx, serviceURL, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{http.StatusOK},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
