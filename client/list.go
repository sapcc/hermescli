// Copyright 2020 SAP SE
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

package client

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/olekukonko/tablewriter"
	"github.com/sapcc/gophercloud-sapcc/audit/v1/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxOffset = 10000

// precision of the overlap detection
const precision = 100

var defaultListKeyOrder = []string{
	"ID",
	"Time",
	"Source",
	"Action",
	"Outcome",
	"RequestPath",
	"Target",
	"Initiator",
}

func parseTime(timeStr string) (time.Time, error) {
	validTimeFormats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04:05-0700"}
	var t time.Time
	var err error
	for _, timeFormat := range validTimeFormats {
		t, err = time.Parse(timeFormat, timeStr)
		if err == nil {
			return t, nil
		}
	}
	return time.Now(), err
}

func getTimeListOpts(allEvents *[]events.Event, listOpts *events.ListOpts) error {
	// time of the last event
	t := (*allEvents)[len(*allEvents)-1]
	rt, err := parseTime(t.EventTime)
	if err != nil {
		return fmt.Errorf("failed to parse time of the last %s event: %s", t.ID, err)
	}

	var filter events.DateFilter
	if getTimeSort(*listOpts) {
		filter = events.DateFilterLTE
	} else {
		filter = events.DateFilterGTE
	}

	dateFilter := events.DateQuery{
		Date:   rt,
		Filter: filter,
	}

	if len(listOpts.Time) > 0 {
		var found bool
		for i, v := range listOpts.Time {
			if v.Filter == filter {
				found = true
				listOpts.Time[i].Date = rt
			}
		}
		if !found {
			listOpts.Time = append(listOpts.Time, dateFilter)
		}
	} else {
		listOpts.Time = []events.DateQuery{
			dateFilter,
		}
	}

	return nil
}

func getNextOffset(page pagination.Page) (int, error) {
	// detect next URL offset
	next, err := page.NextPageURL()
	if err != nil {
		return 0, fmt.Errorf("failed to detect next page url: %s", err)
	}
	parsedURL, err := url.Parse(next)
	if err != nil {
		return 0, fmt.Errorf("failed to parse next url: %s", err)
	}
	params := parsedURL.Query()
	if v, ok := params["offset"]; ok {
		if len(v) == 0 || len(v) > 1 {
			return 0, fmt.Errorf("failed to detect offset")
		}
		return strconv.Atoi(v[0])
	}
	return 0, nil
}

func getTimeSort(listOpts events.ListOpts) bool {
	for _, v := range strings.Split(listOpts.Sort, ",") {
		s := strings.SplitN(v, ":", 2)
		if len(s) == 2 && s[0] == "time" {
			if s[1] == "asc" {
				return false
			}
			if s[1] == "desc" {
				return true
			}
			return false
		}
		if s[0] == "time" {
			return false
		}
	}
	return true
}

func getEvents(client *gophercloud.ServiceClient, allEvents *[]events.Event, listOpts events.ListOpts, userLimit int, precise bool, bar **pb.ProgressBar) error {
	var forceWorkaround bool
	var eventLength int

	err := events.List(client, listOpts).EachPage(func(page pagination.Page) (bool, error) {
		evnts, err := events.ExtractEvents(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract events: %s", err)
		}

		if precise {
			// add only unique events
			// detect duplicates of only previous 100 last and further 100 first items
			// otherwise it is very slow for an amount of objects > 10000 (10000^2*pageN iterations)
			eventLength = len(*allEvents)
		ROOTLOOP:
			for i, evntNew := range evnts {
				for k, j := 0, eventLength-1; j >= eventLength-precision && j >= 0; k, j = k+1, j-1 {
					if k >= precision {
						// don't compare items above 100, break the loop
						break
					}
					if (*allEvents)[j].ID == evntNew.ID {
						continue ROOTLOOP
					}
				}
				if i >= precision {
					// append all remaining and exit the loop
					*allEvents = append(*allEvents, evnts[i:]...)
					break
				} else {
					*allEvents = append(*allEvents, evntNew)
				}
			}
		} else {
			*allEvents = append(*allEvents, evnts...)
		}

		eventLength = len(*allEvents)

		if *bar == nil {
			if v, err := page.(events.EventPage).Total(); err != nil {
				return false, fmt.Errorf("failed to extract total: %s", err)
			} else if eventLength <= maxOffset && eventLength != userLimit {
				if userLimit >= maxOffset && v > userLimit {
					*bar = pb.New(userLimit)
				} else if v > maxOffset {
					*bar = pb.New(v)
				}
				if *bar != nil {
					(*bar).Start()
				}
			}
		}

		if *bar != nil {
			(*bar).Set("Events", eventLength)
		}

		if userLimit > 0 && eventLength >= userLimit {
			// break the loop, when output userLimit is reached
			return false, nil
		}

		nextOffset, err := getNextOffset(page)
		if err != nil {
			return false, err
		}

		if (userLimit == 0 || userLimit > maxOffset) && nextOffset >= maxOffset {
			// go to the workaround to avoid the 500 http code
			forceWorkaround = true
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to list events: %s", err)
	}

	if forceWorkaround && eventLength > 0 {
		// workaround to avoid 10000 limit 500 code
		if err = getTimeListOpts(allEvents, &listOpts); err != nil {
			return err
		}
		delta := userLimit - eventLength
		if delta > 0 && delta <= maxOffset {
			listOpts.Limit = delta
		}
		return getEvents(client, allEvents, listOpts, userLimit, precise, bar)
	}

	return nil
}

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List Hermes events",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		// check time flag
		teq := viper.GetString("time")
		tgt := viper.GetString("time-start")
		tlt := viper.GetString("time-end")
		if teq != "" && !(tgt == "" && tlt == "") {
			return fmt.Errorf("cannot combine time flag with time-start or time-end flags")
		}

		return verifyGlobalFlags(defaultListKeyOrder)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// list events

		userLimit := viper.GetInt("limit")
		keyOrder := viper.GetStringSlice("column")
		if len(keyOrder) == 0 {
			keyOrder = defaultListKeyOrder
		}
		format := viper.GetString("format")

		projectID := viper.GetString("project-id")
		if viper.GetBool("all-projects") {
			projectID = "*"
		}

		listOpts := events.ListOpts{
			Limit:         maxOffset,
			TargetType:    viper.GetString("target-type"),
			TargetID:      viper.GetString("target-id"),
			InitiatorID:   viper.GetString("initiator-id"),
			InitiatorName: viper.GetString("initiator-name"),
			Action:        viper.GetString("action"),
			Outcome:       viper.GetString("outcome"),
			RequestPath:   viper.GetString("request-path"),
			ObserverType:  viper.GetString("source"),
			ProjectID:     projectID,
			Sort:          strings.Join(viper.GetStringSlice("sort"), ","),
		}

		// handle user limits <= 10000
		if userLimit > 0 && userLimit <= maxOffset {
			// default per page limit
			listOpts.Limit = userLimit
		}

		if t := viper.GetString("time"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("failed to parse time: %s", err)
			}
			listOpts.Time = []events.DateQuery{
				{
					Date: rt,
				},
			}
		}
		if t := viper.GetString("time-start"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("failed to parse time-start: %s", err)
			}
			listOpts.Time = append(listOpts.Time, events.DateQuery{
				Date:   rt,
				Filter: events.DateFilterGTE,
			})
		}
		if t := viper.GetString("time-end"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("failed to parse time-end: %s", err)
			}
			listOpts.Time = append(listOpts.Time, events.DateQuery{
				Date:   rt,
				Filter: events.DateFilterLTE,
			})
		}

		client, err := NewHermesV1Client()
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %s", err)
		}

		var allEvents []events.Event
		var bar *pb.ProgressBar

		if err = getEvents(client, &allEvents, listOpts, userLimit, viper.GetBool("over-10k-fix"), &bar); err != nil {
			if bar != nil {
				bar.Finish()
			}
			return fmt.Errorf("failed to list the events: %s", err)
		}
		if bar != nil {
			bar.Finish()
		}

		if format == "table" {
			var buf bytes.Buffer
			table := tablewriter.NewWriter(&buf)
			table.SetColWidth(20)
			table.SetAlignment(3)
			table.SetHeader(keyOrder)

			for _, v := range allEvents {
				kv := eventToKV(v)
				tableRow := []string{}
				for _, k := range keyOrder {
					v := kv[k]
					tableRow = append(tableRow, v)
				}
				table.Append(tableRow)
			}

			table.Render()

			fmt.Print(buf.String())
		} else {
			return printEvent(allEvents, format, keyOrder)
		}

		return nil
	},
}

func init() {
	initListCmdFlags()
	RootCmd.AddCommand(ListCmd)
}

func initListCmdFlags() {
	ListCmd.Flags().StringP("target-type", "", "", "filter events by a target type")
	ListCmd.Flags().StringP("target-id", "", "", "filter events by a target ID")
	ListCmd.Flags().StringP("initiator-id", "", "", "filter events by an initiator ID")
	ListCmd.Flags().StringP("initiator-name", "", "", "filter events by an initiator name")
	ListCmd.Flags().StringP("action", "", "", "filter events by an action")
	ListCmd.Flags().StringP("outcome", "", "", "filter events by an outcome")
	ListCmd.Flags().StringP("request-path", "", "", "filter events by a request path")
	ListCmd.Flags().StringP("source", "", "", "filter events by a source")
	ListCmd.Flags().StringP("time", "", "", "filter events by time")
	ListCmd.Flags().StringP("time-start", "", "", "filter events from time")
	ListCmd.Flags().StringP("time-end", "", "", "filter events till time")
	ListCmd.Flags().StringP("project-id", "", "", "filter events by the project or domain ID (admin only)")
	ListCmd.Flags().BoolP("all-projects", "A", false, "include all projects and domains (admin only) (alias for --project-id '*')")
	ListCmd.Flags().BoolP("over-10k-fix", "", true, "workaround to filter out overlapping events for > 10k total events")
	ListCmd.Flags().UintP("limit", "l", 0, "limit an amount of events in output")
	ListCmd.Flags().StringSliceP("sort", "s", []string{}, `supported sort keys include time, observer_type, target_type, target_id, initiator_type, initiator_id, outcome and action
each sort key may also include a direction suffix
supported directions are ":asc" for ascending and ":desc" for descending
can be specified multiple times`)
}
