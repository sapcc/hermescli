package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	fake "github.com/gophercloud/gophercloud/testhelper/client"
	"github.com/sapcc/hermes-ctl/audit/v1/events"
	"github.com/sapcc/hermes/pkg/cadf"
)

var eventsList = []events.Event{
	{
		ID:        "d3f6695e-8a55-5db1-895c-9f7f0910b7a5",
		EventTime: "2017-11-01T12:28:58.660965+00:00",
		Action:    "create/role_assignment",
		Outcome:   "success",
		Initiator: cadf.Resource{
			TypeURI: "service/security/account/user",
			ID:      "21ff350bc75824262c60adfc58b7fd4a7349120b43a990c2888e6b0b88af6398",
		},
		Target: cadf.Resource{
			TypeURI: "service/security/account/user",
			ID:      "c4d3626f405b99f395a1c581ed630b2d40be8b9701f95f7b8f5b1e2cf2d72c1b",
		},
		Observer: cadf.Resource{
			TypeURI: "service/security",
			ID:      "0e8a00bf-e36c-5a51-9418-2d56d59c8887",
		},
	},
}

var event = events.Event{
	TypeURI:     "http://schemas.dmtf.org/cloud/audit/1.0/event",
	ID:          "7189ce80-6e73-5ad9-bdc5-dcc47f176378",
	EventTime:   "2017-12-18T18:27:32.352893+00:00",
	Action:      "create",
	EventType:   "activity",
	Outcome:     "success",
	RequestPath: "/v2.0/ports.json",
	Reason: cadf.Reason{
		ReasonCode: "201",
		ReasonType: "HTTP",
	},
	Initiator: cadf.Resource{
		TypeURI:   "service/security/account/user",
		ID:        "ba8304b657fb4568addf7116f41b4a16",
		Name:      "neutron",
		Domain:    "Default",
		ProjectID: "ba8304b657fb4568addf7116f41b4a16",
		Host: &cadf.Host{
			Address: "127.0.0.1",
			Agent:   "python-neutronclient",
		},
	},
	Target: cadf.Resource{
		TypeURI:   "network/port",
		ID:        "7189ce80-6e73-5ad9-bdc5-dcc47f176378",
		ProjectID: "ba8304b657fb4568addf7116f41b4a16",
	},
	Observer: cadf.Resource{
		TypeURI: "service/network",
		Name:    "neutron",
		ID:      "7189ce80-6e73-5ad9-bdc5-dcc47f176378",
	},
}

func TestList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, ListResponse)
	})

	count := 0

	events.List(fake.ServiceClient(), events.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := events.ExtractEvents(page)
		if err != nil {
			t.Errorf("Failed to extract events: %v", err)
			return false, nil
		}

		th.CheckDeepEquals(t, eventsList, actual)

		return true, nil
	})

	if count != 1 {
		t.Errorf("Expected 1 page, got %d", count)
	}
}

func TestListOpts(t *testing.T) {
	// Detail cannot take Fields
	opts := events.ListOpts{
		Action:  "create/role_assignment",
		Outcome: "success",
	}

	// Regular ListOpts can
	query, err := opts.ToEventListQuery()
	th.AssertEquals(t, query, "?action=create%2Frole_assignment&outcome=success")
	th.AssertNoErr(t, err)
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/events/7189ce80-6e73-5ad9-bdc5-dcc47f176378", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, GetResponse)
	})

	n, err := events.Get(fake.ServiceClient(), "7189ce80-6e73-5ad9-bdc5-dcc47f176378", nil).Extract()
	th.AssertNoErr(t, err)

	th.AssertDeepEquals(t, *n, event)
}
