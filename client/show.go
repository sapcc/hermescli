package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/kayrus/gophercloud-hermes/audit/v1/events"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ShowCmd represents the show command
var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "short3",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// check required event id
		if len(viper.GetString("id")) == 0 {
			return fmt.Errorf("no event id given")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// show event

		client, err := NewHermesV1Client()
		if err != nil {
			log.Fatal(err)
		}

		// create table
		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		table.SetColWidth(20)
		table.SetAlignment(3)
		table.SetHeader([]string{"ID", "Time", "Source", "Action", "Outcome", "Target", "Initiator"})

		event, err := events.Get(client, viper.GetString("id")).Extract()
		if err != nil {
			log.Fatal(err)
		}

		tableRow := []string{}
		tableRow = append(tableRow, fmt.Sprintf("%v", event.ID))
		tableRow = append(tableRow, fmt.Sprintf("%v", event.EventTime.Format("2006-01-02T15:04:05-0700")))
		tableRow = append(tableRow, fmt.Sprintf("%v", event.Observer.TypeURI))
		tableRow = append(tableRow, fmt.Sprintf("%v", event.Action))
		tableRow = append(tableRow, fmt.Sprintf("%v", event.Outcome))
		tableRow = append(tableRow, fmt.Sprintf("%v\n%v", event.Target.TypeURI, event.Target.ID))
		tableRow = append(tableRow, fmt.Sprintf("%v", event.Initiator.Name))
		table.Append(tableRow)

		// print out
		table.Render()

		fmt.Print(buf.String())

		return nil
	},
}

func init() {
	RootCmd.AddCommand(ShowCmd)
	initShowCmdFlags()

}

func initShowCmdFlags() {
	ShowCmd.Flags().StringP("id", "", "", "id description")
	viper.BindPFlag("id", ShowCmd.Flags().Lookup("id"))
}
