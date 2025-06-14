// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultShowKeyOrder = []string{
	"ID",
	"Type",
	"Time",
	"Observer",
	"TypeURI",
	"Action",
	"Outcome",
	"Target",
	"Initiator",
	"InitiatorDomain",
	"InitiatorAddress",
	"InitiatorAgent",
	"InitiatorAppCredential",
	"RequestPath",
	"Attachments",
}

// ShowCmd represents the show command
var ShowCmd = &cobra.Command{
	Use:   "show <event-id> [<event-id>...]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Show Hermes event",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		return verifyGlobalFlags(defaultShowKeyOrder)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// show event
		client, err := NewHermesV1Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		keyOrder := viper.GetStringSlice("column")
		if len(keyOrder) == 0 {
			keyOrder = defaultShowKeyOrder
		}
		format := viper.GetString("format")

		// initialize the progress bar, when multiple events are requested
		var bar *pb.ProgressBar
		if len(args) > 1 {
			bar = pb.New(len(args))
			bar.SetWriter(os.Stderr)
			bar.Start()
		}

		projectID := viper.GetString("project-id")
		if viper.GetBool("all-projects") {
			projectID = "*"
		}

		getOpts := events.GetOpts{
			ProjectID: projectID,
		}

		var allEvents []events.Event
		for i, id := range args {
			if bar != nil {
				bar.SetCurrent(int64(i + 1))
			}
			event, err := events.Get(cmd.Context(), client, id, getOpts).Extract()
			if err != nil {
				log.Printf("[WARNING] Failed to get %s event: %s", id, err)
				continue
			}
			allEvents = append(allEvents, *event)
		}

		// stop the progress bar
		if bar != nil {
			bar.Finish()
		}

		if format == "table" {
			for _, event := range allEvents {
				kv := eventToKV(event)

				// create table
				var buf bytes.Buffer
				table := tablewriter.NewTable(&buf, tablewriter.WithColumnMax(20), tablewriter.WithRowAlignment(tw.AlignRight))
				table.Header("Key", "Value")

				// populate output table
				for _, k := range keyOrder {
					if v, ok := kv[k]; ok {
						if err := table.Append(k, v); err != nil {
							log.Printf("Error appending row for key %s: %v", k, err)
						}
					}
				}

				if err := table.Render(); err != nil {
					log.Printf("Error rendering table for event %s: %v", event.ID, err)
				}

				fmt.Print(buf.String())
			}
		} else {
			return printEvent(allEvents, format, keyOrder)
		}

		return nil
	},
}

func init() {
	initShowCmdFlags()
	RootCmd.AddCommand(ShowCmd)
}

func initShowCmdFlags() {
	ShowCmd.Flags().StringP("project-id", "", "", "show event for the project or domain ID (admin only)")
	ShowCmd.Flags().BoolP("all-projects", "A", false, "include all projects and domains (admin only) (alias for --project-id '*')")
}
