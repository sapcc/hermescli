package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/pagination"
	"github.com/kayrus/gophercloud-hermes/audit/v1/events"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Hermes events",
	RunE: func(cmd *cobra.Command, args []string) error {
		// list events

		client, err := NewHermesV1Client()
		if err != nil {
			return fmt.Errorf("Failed to create Hermes client: %s", err)
		}

		limit := viper.GetInt("limit")

		listOpts := events.ListOpts{
			Limit:         limit,
			TargetType:    viper.GetString("target-type"),
			InitiatorName: viper.GetString("initiator-name"),
			Action:        viper.GetString("action"),
			Outcome:       viper.GetString("outcome"),
			ObserverType:  viper.GetString("source"),
			// TODO: verify why only time sort works in hermes server
			Sort: strings.Join(viper.GetStringSlice("sort"), ","),
		}

		if limit == 0 {
			// default per page limit
			listOpts.Limit = 5000
		}

		var allEvents []events.Event

		if t := viper.GetString("time"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("Failed to parse time: %s", err)
			}
			listOpts.Time = []events.DateQuery{
				{
					Date: rt,
				},
			}
		} else {
			if t := viper.GetString("time-start"); t != "" {
				rt, err := parseTime(t)
				if err != nil {
					return fmt.Errorf("Failed to parse time-start: %s", err)
				}
				listOpts.Time = append(listOpts.Time, events.DateQuery{
					Date:   rt,
					Filter: events.DateFilterGTE,
				})
			}
			if t := viper.GetString("time-end"); t != "" {
				rt, err := parseTime(t)
				if err != nil {
					return fmt.Errorf("Failed to parse time-end: %s", err)
				}
				listOpts.Time = append(listOpts.Time, events.DateQuery{
					Date:   rt,
					Filter: events.DateFilterLTE,
				})
			}
		}

		err = events.List(client, listOpts).EachPage(func(page pagination.Page) (bool, error) {
			evnts, err := events.ExtractEvents(page)
			if err != nil {
				return false, fmt.Errorf("Failed to extract events: %s", err)
			}

			allEvents = append(allEvents, evnts...)

			if limit > 0 && len(allEvents) >= limit {
				// break the loop
				return false, nil
			}

			return true, nil
		})
		if err != nil {
			return fmt.Errorf("Failed to list events: %s", err)
		}

		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		table.SetColWidth(20)
		table.SetAlignment(3)
		table.SetHeader([]string{"ID", "Time", "Source", "Action", "Outcome", "Target", "Initiator"})

		for _, v := range allEvents {
			tableRow := []string{}
			tableRow = append(tableRow, fmt.Sprintf("%s", v.ID))
			tableRow = append(tableRow, fmt.Sprintf("%s", v.EventTime.Format("2006-01-02T15:04:05-0700")))
			tableRow = append(tableRow, fmt.Sprintf("%s", v.Observer.TypeURI))
			tableRow = append(tableRow, fmt.Sprintf("%s", v.Action))
			tableRow = append(tableRow, fmt.Sprintf("%s", v.Outcome))
			tableRow = append(tableRow, fmt.Sprintf("%s\n%s", v.Target.TypeURI, v.Target.ID))
			tableRow = append(tableRow, fmt.Sprintf("%s", v.Initiator.Name))
			table.Append(tableRow)
		}

		// print out
		table.Render()

		fmt.Print(buf.String())

		return nil
	},
}

func init() {
	initListCmdFlags()
	RootCmd.AddCommand(ListCmd)
}

func initListCmdFlags() {
	ListCmd.Flags().StringP("target-type", "", "", "filter events by a target type")
	ListCmd.Flags().StringP("initiator-name", "", "", "filter events by an initiator name")
	ListCmd.Flags().StringP("action", "", "", "filter events by an action")
	ListCmd.Flags().StringP("outcome", "", "", "filter events by an outcome")
	ListCmd.Flags().StringP("source", "", "", "filter events by a source")
	// TODO: add conflict with the time and time-start/time-end
	ListCmd.Flags().StringP("time", "", "", "filter events by time")
	ListCmd.Flags().StringP("time-start", "", "", "filter events from time")
	ListCmd.Flags().StringP("time-end", "", "", "filter events till time")
	ListCmd.Flags().IntP("limit", "", 0, "limit an amount of events in output")
	ListCmd.Flags().StringSliceP("sort", "", []string{}, `supported sort keys include time, observer_type, target_type, target_id, initiator_type, initiator_id, outcome and action
each sort key may also include a direction suffix
supported directions are ":asc" for ascending and ":desc" for descending
can be specified multiple times`)
	viper.BindPFlag("initiator-name", ListCmd.Flags().Lookup("initiator-name"))
	viper.BindPFlag("target-type", ListCmd.Flags().Lookup("target-type"))
	viper.BindPFlag("action", ListCmd.Flags().Lookup("action"))
	viper.BindPFlag("outcome", ListCmd.Flags().Lookup("outcome"))
	viper.BindPFlag("source", ListCmd.Flags().Lookup("source"))
	viper.BindPFlag("time", ListCmd.Flags().Lookup("time"))
	viper.BindPFlag("time-start", ListCmd.Flags().Lookup("time-start"))
	viper.BindPFlag("time-end", ListCmd.Flags().Lookup("time-end"))
	viper.BindPFlag("limit", ListCmd.Flags().Lookup("limit"))
	viper.BindPFlag("sort", ListCmd.Flags().Lookup("sort"))
}
