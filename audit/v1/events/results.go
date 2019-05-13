package events

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/sapcc/hermes/pkg/cadf"
)

type JSONRFC3339 time.Time

func (jt *JSONRFC3339) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339(t)
	return nil
}

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
type Event struct {
	// CADF Event Schema
	TypeURI string `json:"typeURI"`

	// CADF generated event id
	ID string `json:"id"`

	// CADF generated timestamp
	EventTime time.Time `json:"-"`

	// Characterizes events: eg. activity
	EventType string `json:"eventType"`

	// CADF action mapping for GET call on an OpenStack REST API
	Action string `json:"action"`

	// Outcome of REST API call, eg. success/failure
	Outcome string `json:"outcome"`

	// Standard response for successful HTTP requests
	Reason cadf.Reason `json:"reason,omitempty"`

	// CADF component that contains the RESOURCE
	// that initiated, originated, or instigated the event's
	// ACTION, according to the OBSERVER
	Initiator cadf.Resource `json:"initiator"`

	// CADF component that contains the RESOURCE
	// against which the ACTION of a CADF Event
	// Record was performed, was attempted, or is
	// pending, according to the OBSERVER.
	Target cadf.Resource `json:"target"`

	// CADF component that contains the RESOURCE
	// that generates the CADF Event Record based on
	// its observation (directly or indirectly) of the Actual Event
	Observer cadf.Resource `json:"observer"`

	// Attachment contains self-describing extensions to the event
	Attachments []cadf.Attachment `json:"attachments,omitempty"`

	// Request path on the OpenStack service REST API call
	RequestPath string `json:"requestPath,omitempty"`
}

func (r *Event) UnmarshalJSON(b []byte) error {
	type tmp Event
	var s struct {
		tmp
		EventTime JSONRFC3339 `json:"eventTime"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Event(s.tmp)

	r.EventTime = time.Time(s.EventTime)

	return nil
}

func (r *Event) MarshalJSON() ([]byte, error) {
	type ext struct {
		EventTime string `json:"eventTime"`
	}
	type tmp struct {
		ext
		Event
	}

	s := tmp{
		ext{
			EventTime: r.EventTime.Format(time.RFC3339),
		},
		Event{
			ID:          r.ID,
			TypeURI:     r.TypeURI,
			Attachments: r.Attachments,
			Action:      r.Action,
			EventType:   r.EventType,
			RequestPath: r.RequestPath,
			Reason:      r.Reason,
			Outcome:     r.Outcome,
			Initiator:   r.Initiator,
			Target:      r.Target,
			Observer:    r.Observer,
		},
	}

	return json.Marshal(s)
}

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
