// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/url"
	"testing"

	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/events"
)

func TestDetailedListOpts(t *testing.T) {
	base := events.ListOpts{Search: `"2042"`, Limit: 10}

	q, err := detailedListOpts{base, false}.ToEventListQuery()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(q)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Has("details") {
		t.Errorf("expected no details parameter, got %s", q)
	}

	q, err = detailedListOpts{base, true}.ToEventListQuery()
	if err != nil {
		t.Fatal(err)
	}
	u, err = url.Parse(q)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("details") != "true" || u.Query().Get("search") != `"2042"` || u.Query().Get("limit") != "10" {
		t.Errorf("expected details, search and limit parameters, got %s", q)
	}
}
