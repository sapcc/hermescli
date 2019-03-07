// +build acceptance networking

package audit

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/kayrus/gophercloud-hermes/audit/v1/events"
)

func TestEventList(t *testing.T) {
	client, err := NewHermesV1Client()
	th.AssertNoErr(t, err)

	var count int
	var allEvents []events.Event

	events.List(client, events.ListOpts{Limit: 5000}).EachPage(func(page pagination.Page) (bool, error) {
		count++
		tmp, err := events.ExtractEvents(page)
		if err != nil {
			t.Errorf("Failed to extract events: %v", err)
			return false, nil
		}

		allEvents = append(allEvents, tmp...)

		return true, nil
	})

	tools.PrintResource(t, allEvents)

	expectedPages := 2
	if count < expectedPages {
		t.Errorf("Expected %d page, got %d", expectedPages, count)
	}

	expectedEvents := 2
	if len(allEvents) < expectedEvents {
		t.Errorf("Expected %d events, got %d", expectedEvents, len(allEvents))
	}
}
