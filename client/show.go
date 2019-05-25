package client

import (
	"bytes"
	"fmt"
	"log"

	"github.com/olekukonko/tablewriter"
	"github.com/sapcc/hermes-ctl/audit/v1/events"
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
	"RequestPath",
	"Attachments",
}

// ShowCmd represents the show command
var ShowCmd = &cobra.Command{
	Use:   "show <event-id> [<event-id>...]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Show Hermes event",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return verifyGlobalFlags(defaultShowKeyOrder)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// show event
		client, err := NewHermesV1Client()
		if err != nil {
			return fmt.Errorf("Failed to create Hermes client: %s", err)
		}

		keyOrder := viper.GetStringSlice("column")
		if len(keyOrder) == 0 {
			keyOrder = defaultShowKeyOrder
		}
		format := viper.GetString("format")

		var allEvents []events.Event
		for _, id := range args {
			event, err := events.Get(client, id).Extract()
			if err != nil {
				log.Printf("[WARNING] Failed to get %s event: %s", id, err)
				continue
			}
			allEvents = append(allEvents, *event)
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
	RootCmd.AddCommand(ShowCmd)
}
