package client

import (
	"testing"

	"github.com/sapcc/go-api-declarations/cadf"
	"github.com/sapcc/gophercloud-sapcc/audit/v1/events"
)

func TestEventToKV(t *testing.T) {
	event := events.Event{
		ID:        "1",
		EventType: "TestEvent",
		EventTime: "2023-06-14",
		Observer: cadf.Resource{
			Name:    "Observer1",
			TypeURI: "ObserverTypeURI",
		},
		Action:  cadf.Action("Action"),
		Outcome: cadf.Outcome("Outcome"),
		Target:  cadf.Resource{TypeURI: "TargetTypeURI", ID: "TargetID"},
		Initiator: cadf.Resource{
			Name:   "InitiatorName",
			Domain: "InitiatorDomain",
			Host: &cadf.Host{
				Address: "InitiatorAddress",
				Agent:   "InitiatorAgent",
			},
			AppCredentialID: "InitiatorAppCredentialID",
		},
		RequestPath: "RequestPath",
	}

	expected := map[string]string{
		"ID":                     "1",
		"Type":                   "TestEvent",
		"Time":                   "2023-06-14",
		"Observer":               "Observer1",
		"TypeURI":                "ObserverTypeURI",
		"Source":                 "ObserverTypeURI",
		"Action":                 "Action",
		"Outcome":                "Outcome",
		"Target":                 "TargetTypeURI TargetID",
		"Initiator":              "InitiatorName",
		"InitiatorDomain":        "InitiatorDomain",
		"InitiatorAddress":       "InitiatorAddress",
		"InitiatorAgent":         "InitiatorAgent",
		"InitiatorAppCredential": "InitiatorAppCredentialID",
		"RequestPath":            "RequestPath",
	}

	result := eventToKV(event)

	for k, v := range expected {
		if result[k] != v {
			t.Errorf("expected key %s to be %s but got %s", k, v, result[k])
		}
	}
}
