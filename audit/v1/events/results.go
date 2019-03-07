package events

import (
	"encoding/json"
	"time"
	//"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// TODO: validTimeFormats := []string{time.RFC3339, "2006-01-02T15:04:05-0700", "2006-01-02T15:04:05"}

// RFC3339Z is the time format used in Zun (Containers Service).
const RFC3339Z = "2006-01-02T15:04:05-07:00"

type JSONRFC3339Z time.Time

func (jt *JSONRFC3339Z) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(RFC3339Z, s)
	//panic(fmt.Sprintf("%v", t))
	if err != nil {
		return err
	}
	*jt = JSONRFC3339Z(t)
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

// TODO: hermes/pkg/cadf/event.go

type Reason struct {
	ReasonCode string `json:"reasonCode,omitempty"`
	ReasonType string `json:"reasonType,omitempty"`
}

type Host struct {
	Address string `json:"address,omitempty"`
	Agent   string `json:"agent,omitempty"`
}

type Initiator struct {
	ID        string `json:"id"`
	TypeURI   string `json:"typeURI"`
	Name      string `json:"name"`
	Domain    string `json:"domain,omitempty"`
	DomainID  string `json:"domain_id,omitempty"`
	Project   string `json:"project,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Host      Host   `json:"host,omitempty"`
}

type Attachment struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

type Target struct {
	ID          string       `json:"id"`
	TypeURI     string       `json:"typeURI"`
	ProjectID   string       `json:"project_id,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Observer struct {
	ID      string `json:"id"`
	TypeURI string `json:"typeURI"`
	Name    string `json:"name"`
}

// Event represents a Hermes Event.
type Event struct {
	ID          string    `json:"id,omitempty"`
	TypeURI     string    `json:"typeURI,omitempty"`
	EventTime   time.Time `json:"-"`
	Action      string    `json:"action,omitempty"`
	EventType   string    `json:"eventType,omitempty"`
	ResourceID  string    `json:"resource_id,omitempty"`
	RequestPath string    `json:"requestPath,omitempty"`
	Reason      Reason    `json:"reason,omitempty"`
	Outcome     string    `json:"outcome,omitempty"`
	Initiator   Initiator `json:"initiator,omitempty"`
	Target      Target    `json:"target,omitempty"`
	Observer    Observer  `json:"observer,omitempty"`
}

func (r *Event) UnmarshalJSON(b []byte) error {
	type tmp Event
	var s struct {
		tmp
		EventTime JSONRFC3339Z `json:"eventTime"`
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
			Action:      r.Action,
			EventType:   r.EventType,
			RequestPath: r.RequestPath,
			ResourceID:  r.ResourceID,
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
