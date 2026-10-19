// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"
	"time"

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
	}

	for _, c := range cases {
		base, err := c.opts.ToEventListQuery()
		if err != nil {
			t.Fatal(err)
		}

		q, err := detailedListOpts{c.opts, false}.ToEventListQuery()
		if err != nil {
			t.Fatal(err)
		}
		if q != base {
			t.Errorf("%s: without details expected %q, got %q", c.name, base, q)
		}

		q, err = detailedListOpts{c.opts, true}.ToEventListQuery()
		if err != nil {
			t.Fatal(err)
		}
		expected := "?details=true"
		if base != "" {
			expected += "&" + base[1:]
		}
		if q != expected {
			t.Errorf("%s: with details expected %q, got %q", c.name, expected, q)
		}
	}
}
