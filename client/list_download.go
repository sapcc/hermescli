// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var listDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download all matching Hermes events to a local file",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		teq := viper.GetString("time")
		tgt := viper.GetString("time-start")
		tlt := viper.GetString("time-end")
		if teq != "" && (tgt != "" || tlt != "") {
			return errors.New("cannot combine time flag with time-start or time-end flags")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFile := viper.GetString("output")

		// --format defaults to "table" (set by the root persistent flag), which has no meaning for
		// file output. Coerce it to "json". "value" is also display-only and is rejected explicitly.
		format := viper.GetString("format")
		switch format {
		case "table", "":
			format = "json"
		case "value":
			return errors.New(`"value" format is not supported for download, use json, yaml, or csv`)
		}

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
			Search:        viper.GetString("search"),
			ProjectID:     projectID,
			Sort:          strings.Join(viper.GetStringSlice("sort"), ","),
		}

		if t := viper.GetString("time"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("failed to parse time: %w", err)
			}
			listOpts.Time = []events.DateQuery{{Date: rt}}
		}
		if t := viper.GetString("time-start"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("failed to parse time-start: %w", err)
			}
			listOpts.Time = append(listOpts.Time, events.DateQuery{
				Date:   rt,
				Filter: events.DateFilterGTE,
			})
		}
		if t := viper.GetString("time-end"); t != "" {
			rt, err := parseTime(t)
			if err != nil {
				return fmt.Errorf("failed to parse time-end: %w", err)
			}
			listOpts.Time = append(listOpts.Time, events.DateQuery{
				Date:   rt,
				Filter: events.DateFilterLTE,
			})
		}

		hermesClient, err := NewHermesV1Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		var allEvents []events.Event
		var bar *pb.ProgressBar
		if err = getEvents(cmd.Context(), hermesClient, &allEvents, listOpts, 0, true, &bar); err != nil {
			if bar != nil {
				bar.Finish()
			}
			return fmt.Errorf("failed to list events: %w", err)
		}
		if bar != nil {
			bar.Finish()
		}

		var out *os.File
		if outputFile == "-" {
			out = os.Stdout
		} else {
			out, err = os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()
		}

		switch format {
		case "json":
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			if err := enc.Encode(allEvents); err != nil {
				return fmt.Errorf("failed to write JSON: %w", err)
			}
		case "yaml":
			enc := yaml.NewEncoder(out)
			if err := enc.Encode(allEvents); err != nil {
				return fmt.Errorf("failed to write YAML: %w", err)
			}
			enc.Close()
		case "csv":
			keyOrder := viper.GetStringSlice("column")
			if len(keyOrder) == 0 {
				keyOrder = defaultListKeyOrder
			}
			if err := writeCSV(out, allEvents, keyOrder); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported format %q, use: json, yaml, csv", format)
		}

		if outputFile != "-" {
			fmt.Fprintf(os.Stderr, "Downloaded %d events to %s\n", len(allEvents), outputFile)
		}

		return nil
	},
}

func init() {
	ListCmd.AddCommand(listDownloadCmd)
	listDownloadCmd.Flags().StringP("output", "o", "events.json", "output file path (use - for stdout)")
	listDownloadCmd.Flags().StringP("target-type", "", "", "filter events by a target type")
	listDownloadCmd.Flags().StringP("target-id", "", "", "filter events by a target ID")
	listDownloadCmd.Flags().StringP("initiator-id", "", "", "filter events by an initiator ID")
	listDownloadCmd.Flags().StringP("initiator-name", "", "", "filter events by an initiator name")
	listDownloadCmd.Flags().StringP("action", "", "", "filter events by an action")
	listDownloadCmd.Flags().StringP("outcome", "", "", "filter events by an outcome")
	listDownloadCmd.Flags().StringP("request-path", "", "", "filter events by a request path")
	listDownloadCmd.Flags().StringP("source", "", "", "filter events by a source")
	listDownloadCmd.Flags().StringP("search", "", "", "filter events by a search string")
	listDownloadCmd.Flags().StringP("time", "", "", "filter events by time")
	listDownloadCmd.Flags().StringP("time-start", "", "", "filter events from time")
	listDownloadCmd.Flags().StringP("time-end", "", "", "filter events till time")
	listDownloadCmd.Flags().StringP("project-id", "", "", "filter events by the project or domain ID (admin only)")
	listDownloadCmd.Flags().BoolP("all-projects", "A", false, "include all projects and domains (admin only) (alias for --project-id '*')")
	listDownloadCmd.Flags().StringSliceP("sort", "s", []string{}, `sort keys: time, observer_type, target_type, target_id, initiator_type, initiator_id, outcome, action (suffix :asc or :desc)`)
}
