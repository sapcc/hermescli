package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/kayrus/gophercloud-hermes/audit/v1/events"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const usage = `Usage:
  hermesctl show [flags] <event-id> [<event-id>...]

Flags:
  -h, --help   help for show

Global Flags:
      --debug   Print out request and response objects.
`

// ShowCmd represents the show command
var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show Hermes event",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// check required event id
		if len(args) == 0 {
			return fmt.Errorf("no event id given")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// show event

		client, err := NewHermesV1Client()
		if err != nil {
			return fmt.Errorf("Failed to create Hermes client: %s", err)
		}

		for _, id := range args {
			event, err := events.Get(client, id).Extract()
			if err != nil {
				log.Printf("[WARNING] %s", err)
				continue
			}

			// create table
			var buf bytes.Buffer
			table := tablewriter.NewWriter(&buf)
			table.SetColWidth(20)
			table.SetAlignment(3)
			table.SetHeader([]string{"Key", "Value"})

			table.Append([]string{"ID", fmt.Sprintf("%s", event.ID)})
			table.Append([]string{"Type", fmt.Sprintf("%s", event.EventType)})
			table.Append([]string{"Time", fmt.Sprintf("%s", event.EventTime.Format("2006-01-02T15:04:05-0700"))})
			if len(event.Observer.Name) > 0 {
				table.Append([]string{"Observer", fmt.Sprintf("%s", event.Observer.Name)})
			}
			table.Append([]string{"TypeURI", fmt.Sprintf("%s", event.Observer.TypeURI)})
			table.Append([]string{"Action", fmt.Sprintf("%s", event.Action)})
			table.Append([]string{"Outcome", fmt.Sprintf("%s", event.Outcome)})
			table.Append([]string{"Target", fmt.Sprintf("%s\n%s", event.Target.TypeURI, event.Target.ID)})
			if len(event.Initiator.Name) > 0 {
				table.Append([]string{"Initiator", fmt.Sprintf("%s", event.Initiator.Name)})
			}
			if len(event.Initiator.Domain) > 0 {
				table.Append([]string{"InitiatorDomain", fmt.Sprintf("%s", event.Initiator.Domain)})
			}
			table.Append([]string{"InitiatorAddress", fmt.Sprintf("%s", event.Initiator.Host.Address)})
			table.Append([]string{"InitiatorAgent", fmt.Sprintf("%s", event.Initiator.Host.Agent)})
			if len(event.RequestPath) > 0 {
				table.Append([]string{"RequestPath", fmt.Sprintf("%s", event.RequestPath)})
			}

			if len(event.Attachments) > 0 {
				var attachments []string
				for _, attachment := range event.Attachments {
					if attachment.Content != nil {
						attachments = append(attachments, fmt.Sprintf("%v", attachment.Content))
					}
				}
				if len(attachments) > 0 {
					table.Append([]string{"Attachments", strings.Join(attachments, "\n")})
				}
			}

			// print out
			table.Render()

			fmt.Print(buf.String())
		}

		return nil
	},
}

func init() {
	ShowCmd.SetUsageTemplate(usage)
	RootCmd.AddCommand(ShowCmd)
	initShowCmdFlags()
}

func initShowCmdFlags() {
	// TODO: add "-c" and "-f" for column and print format
	//ShowCmd.Flags().StringP("id", "", "", "id description")
	//viper.BindPFlag("id", ShowCmd.Flags().Lookup("id"))
}
