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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/sapcc/go-api-declarations/cadf"
)

type GetResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts an event resource.
func (r GetResult) Extract() (*Event, error) {
	var s Event
	err := r.ExtractInto(&s)
	return &s, err
}

func (r GetResult) ExtractInto(v interface{}) error {
	return r.Result.ExtractIntoStructPtr(v, "")
}

// Event represents a Hermes Event.
type Event cadf.Event

// EventPage is a single page of events results.
type EventPage struct {
	pagination.LinkedPageBase
}

// IsEmpty determines whether or not a page of events contains any results.
func (r EventPage) IsEmpty() (bool, error) {
	events, err := ExtractEvents(r)
	return len(events) == 0, err
}

// NextPageURL extracts the "next" link from the links section of the result.
func (r EventPage) NextPageURL() (string, error) {
	var s struct {
		Next     string `json:"next"`
		Previous string `json:"previous"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return s.Next, err
}

// Total extracts the "total" attribute of the result.
func (r EventPage) Total() (int, error) {
	var s struct {
		Total int `json:"total"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return 0, err
	}
	return s.Total, err
}

// ExtractEvents accepts a Page struct, specifically an EventPage struct,
// and extracts the elements into a slice of Event structs. In other words,
// a generic collection is mapped into a relevant slice.
func ExtractEvents(r pagination.Page) ([]Event, error) {
	var s []Event
	err := ExtractEventsInto(r, &s)
	return s, err
}

func ExtractEventsInto(r pagination.Page, v interface{}) error {
	return r.(EventPage).Result.ExtractIntoSlicePtr(v, "events")
}
