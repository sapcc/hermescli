package main

import (
	"bytes"
	"fmt"
	"log"
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
	Short: "short2",
	RunE: func(cmd *cobra.Command, args []string) error {
		// list events

		client, err := NewHermesV1Client()
		if err != nil {
			log.Fatal(err)
		}

		listOpts := events.ListOpts{
			Limit:         5000,
			TargetType:    viper.GetString("target-type"),
			InitiatorName: viper.GetString("initiator-name"),
			Action:        viper.GetString("action"),
			Outcome:       viper.GetString("outcome"),
			ObserverType:  viper.GetString("source"),
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

		events.List(client, listOpts).EachPage(func(page pagination.Page) (bool, error) {
			tmp, err := events.ExtractEvents(page)
			if err != nil {
				log.Fatalf("Failed to extract events: %v", err)
				return false, nil
			}

			allEvents = append(allEvents, tmp...)

			return true, nil
		})

		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		table.SetColWidth(20)
		table.SetAlignment(3)
		table.SetHeader([]string{"ID", "Time", "Source", "Action", "Outcome", "Target", "Initiator"})

		for _, v := range allEvents {
			tableRow := []string{}
			tableRow = append(tableRow, fmt.Sprintf("%v", v.ID))
			tableRow = append(tableRow, fmt.Sprintf("%v", v.EventTime.Format("2006-01-02T15:04:05-0700")))
			tableRow = append(tableRow, fmt.Sprintf("%v", v.Observer.TypeURI))
			tableRow = append(tableRow, fmt.Sprintf("%v", v.Action))
			tableRow = append(tableRow, fmt.Sprintf("%v", v.Outcome))
			tableRow = append(tableRow, fmt.Sprintf("%v\n%v", v.Target.TypeURI, v.Target.ID))
			tableRow = append(tableRow, fmt.Sprintf("%v", v.Initiator.Name))
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
	//initListCmdFlags()
}

func initListCmdFlags() {
	ListCmd.Flags().StringP("target-type", "", "", "target-type description")
	ListCmd.Flags().StringP("initiator-name", "", "", "initiator-name description")
	ListCmd.Flags().StringP("action", "", "", "action description")
	ListCmd.Flags().StringP("outcome", "", "", "outcome description")
	ListCmd.Flags().StringP("source", "", "", "source description")
	// TODO: add conflict with the time and time-start/time-end
	ListCmd.Flags().StringP("time", "", "", "time description")
	ListCmd.Flags().StringP("time-start", "", "", "time-start description")
	ListCmd.Flags().StringP("time-end", "", "", "time-end description")
	viper.BindPFlag("initiator-name", ListCmd.Flags().Lookup("initiator-name"))
	viper.BindPFlag("target-type", ListCmd.Flags().Lookup("target-type"))
	viper.BindPFlag("action", ListCmd.Flags().Lookup("action"))
	viper.BindPFlag("outcome", ListCmd.Flags().Lookup("outcome"))
	viper.BindPFlag("source", ListCmd.Flags().Lookup("source"))
	viper.BindPFlag("time", ListCmd.Flags().Lookup("time"))
	viper.BindPFlag("time-start", ListCmd.Flags().Lookup("time-start"))
	viper.BindPFlag("time-end", ListCmd.Flags().Lookup("time-end"))
}
