// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/events"
)

func TestDetailedListOpts(t *testing.T) {
	ts := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	full := events.ListOpts{
		Search: `"2042"`,
		Limit:  10,
		Time:   []events.DateQuery{{Date: ts, Filter: events.DateFilterGTE}},
	}

	cases := []struct {
		name string
		opts events.ListOpts
	}{
		{"empty", events.ListOpts{}},
		{"search, limit and time", full},
		// action sorts before details
		{"action", events.ListOpts{Action: "create"}},
		{"action and time", events.ListOpts{Action: "create", Time: full.Time}},
	}

	parse := func(t *testing.T, q string) url.Values {
		t.Helper()
		v, err := url.ParseQuery(q)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base, err := c.opts.ToEventListQuery()
			if err != nil {
				t.Fatal(err)
			}
			baseValues := parse(t, base[min(1, len(base)):])

			q, err := detailedListOpts{c.opts, false}.ToEventListQuery()
			if err != nil {
				t.Fatal(err)
			}
			if q != base {
				t.Errorf("without details expected %q, got %q", base, q)
			}

			q, err = detailedListOpts{c.opts, true}.ToEventListQuery()
			if err != nil {
				t.Fatal(err)
			}
			if q == "" || q[0] != '?' {
				t.Fatalf("with details expected a query starting with ?, got %q", q)
			}
			expected := cloneValues(baseValues)
			expected.Set("details", "true")
			if got := parse(t, q[1:]); !reflect.DeepEqual(got, expected) {
				t.Errorf("with details expected %v, got %v", expected, got)
			}
		})
	}
}

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v)+1)
	for k, vals := range v {
		out[k] = slices.Clone(vals)
	}
	return out
}

func TestResolveListColumns(t *testing.T) {
	cases := []struct {
		name        string
		columns     []string
		details     bool
		wantColumns []string
		wantDetails bool
	}{
		{"neither flag", nil, false, defaultListKeyOrder, false},
		{"empty column list", []string{}, false, defaultListKeyOrder, false},
		{"details only", nil, true, listColumns, true},
		{"attachments column only", []string{"Attachments"}, false, []string{"Attachments"}, true},
		{"details with explicit column", []string{"ID"}, true, []string{"ID"}, true},
		{"explicit columns without attachments", []string{"ID", "Action"}, false, []string{"ID", "Action"}, false},
		{"attachments among columns", []string{"ID", "Attachments"}, false, []string{"ID", "Attachments"}, true},
		{"details and attachments column", []string{"Attachments"}, true, []string{"Attachments"}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cols, details := resolveListColumns(c.columns, c.details)
			if !slices.Equal(cols, c.wantColumns) {
				t.Errorf("columns: expected %v, got %v", c.wantColumns, cols)
			}
			if details != c.wantDetails {
				t.Errorf("details: expected %v, got %v", c.wantDetails, details)
			}
		})
	}

	if slices.Contains(defaultListKeyOrder, "Attachments") {
		t.Error("the default columns must not contain Attachments")
	}
	if !slices.Contains(listColumns, "Attachments") {
		t.Error("listColumns must contain Attachments")
	}
}

// TestGetEventsDetailsSurvivesWorkaround checks that details is still sent on
// the follow-up request, which getEvents issues when the 10000 offset limit is
// reached.
func TestGetEventsDetailsSurvivesWorkaround(t *testing.T) {
	for _, details := range []bool{true, false} {
		t.Run(fmt.Sprintf("details=%v", details), func(t *testing.T) {
			var mu sync.Mutex
			var queries []url.Values

			var srv *httptest.Server
			srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				queries = append(queries, r.URL.Query())
				n := len(queries)
				mu.Unlock()

				w.Header().Set("Content-Type", "application/json")
				if n == 1 {
					// one event and a next link pointing at the offset limit
					next := srv.URL + "/events?offset=" + strconv.Itoa(maxOffset)
					fmt.Fprintf(w, `{"events":[{"id":"e1","eventTime":"2026-10-19T10:00:00Z"}],"total":1,"next":%q}`, next)
					return
				}
				fmt.Fprint(w, `{"events":[],"total":0}`)
			}))
			defer srv.Close()

			client := &gophercloud.ServiceClient{
				ProviderClient: &gophercloud.ProviderClient{},
				Endpoint:       srv.URL + "/",
			}

			var all []events.Event
			var bar *pb.ProgressBar
			err := getEvents(t.Context(), client, &all, events.ListOpts{Limit: maxOffset}, details, 0, false, &bar)
			if err != nil {
				t.Fatal(err)
			}

			mu.Lock()
			defer mu.Unlock()
			if len(queries) != 2 {
				t.Fatalf("expected the initial and the workaround request, got %d requests", len(queries))
			}
			if len(all) != 1 || all[0].ID != "e1" {
				t.Errorf("unexpected events: %v", all)
			}
			if queries[1].Get("time") == "" {
				t.Errorf("second request is not the time workaround: %v", queries[1])
			}
			for i, q := range queries {
				got, ok := q["details"]
				switch {
				case details && (!ok || len(got) != 1 || got[0] != "true"):
					t.Errorf("request %d: expected details=true, got %v", i+1, q)
				case !details && ok:
					t.Errorf("request %d: expected no details, got %v", i+1, q)
				}
			}
		})
	}
}
