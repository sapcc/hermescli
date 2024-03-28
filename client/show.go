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
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/sapcc/gophercloud-sapcc/audit/v1/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/cheggaaa/pb.v1"
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
		client, err := NewHermesV1Client()
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
			bar.Output = os.Stderr
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
				bar.Set(i + 1)
			}
			event, err := events.Get(client, id, getOpts).Extract()
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
				table := tablewriter.NewWriter(&buf)
				table.SetColWidth(20)
				table.SetAlignment(3)
				table.SetHeader([]string{"Key", "Value"})

				// populate output table
				for _, k := range keyOrder {
					if v, ok := kv[k]; ok {
						table.Append([]string{k, v})
					}
				}

				table.Render()

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
