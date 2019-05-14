package events

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/sapcc/hermes/pkg/cadf"
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
