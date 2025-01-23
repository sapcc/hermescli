// Copyright 2025 SAP SE
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatYAML ExportFormat = "yaml"
	ExportFormatCSV  ExportFormat = "csv"
)

var (
	allExportFormats = []ExportFormat{
		ExportFormatJSON,
		ExportFormatYAML,
		ExportFormatCSV,
	}
)

func parseExportFormat(input string) (ExportFormat, error) {
	if slices.Contains(allExportFormats, ExportFormat(input)) {
		return ExportFormat(input), nil
	}
	return "", fmt.Errorf("unsupported format: %s (supported formats: %v)", input, allExportFormats)
}

// convertToRequestedFormat converts events to the specified format and writes to the provided buffer
func convertToRequestedFormat(buf *bytes.Buffer, allEvents []events.Event, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(allEvents, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		buf.Write(data)

	case "yaml":
		data, err := yaml.Marshal(allEvents)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		buf.Write(data)

	case "csv":
		if err := writeCSV(buf, allEvents, defaultListKeyOrder); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// ExportCmd represents the export command
var ExportCmd = &cobra.Command{
	Use:   "export",
	Args:  cobra.ExactArgs(0),
	Short: "Export Hermes events to Swift",
	Long: `Export Hermes events to Swift storage container.
Exports can be saved in different formats (json, csv, yaml) for further processing or archival.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}

		// Validate container name is provided
		if viper.GetString("container") == "" {
			return errors.New("container name is required")
		}

		// Validate format
		_, err := parseExportFormat(viper.GetString("format"))
		if err != nil {
			return err
		}

		return verifyGlobalFlags(defaultListKeyOrder)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Warn if trying to export more than default limit
		if viper.GetInt("limit") > maxOffset {
			fmt.Fprintf(os.Stderr, "Warning: Exporting more than %d events may take a long time.\n\n", maxOffset)
		}

		// Get events using existing list functionality
		client, err := NewHermesV1Client(ctx)
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		// Configure debug logging if enabled
		if viper.GetBool("debug") {
			transport := &http.Transport{}
			client.HTTPClient = http.Client{
				Transport: transport,
			}
		}

		fmt.Fprintf(os.Stderr, "Fetching events...\n")

		var allEvents []events.Event
		var bar *pb.ProgressBar

		logg.Debug("fetching events matching specified criteria")

		listOpts := buildListOpts()
		if err = getEvents(ctx, client, &allEvents, listOpts, viper.GetInt("limit"), true, &bar); err != nil {
			if bar != nil {
				bar.Finish()
			}
			return fmt.Errorf("failed to list events: %w", err)
		}
		if bar != nil {
			bar.Finish()
		}

		if len(allEvents) == 0 {
			return errors.New("no events found matching the specified criteria")
		}

		fmt.Fprintf(os.Stderr, "\nFound %d events to export\n", len(allEvents))

		// Convert events to desired format
		fmt.Fprintf(os.Stderr, "Converting to %s format...\n", viper.GetString("format"))
		var buf bytes.Buffer
		if err = convertToRequestedFormat(&buf, allEvents, viper.GetString("format")); err != nil {
			return fmt.Errorf("failed to convert events: %w", err)
		}

		dataSize := float64(buf.Len()) / 1024 / 1024 // Convert to MB
		fmt.Fprintf(os.Stderr, "Uploading %.1fMB to Swift...\n", dataSize)

		// Create upload progress bar
		uploadBar := pb.Full.Start64(int64(buf.Len()))
		uploadBar.Set(pb.Bytes, true)
		uploadBar.SetWidth(80)
		defer uploadBar.Finish()

		// Wrap the buffer in a progress reader
		progressReader := &progressReader{
			Reader: &buf,
			Bar:    uploadBar,
		}

		// Initialize Swift container
		container, err := InitializeSwiftContainer(
			ctx,
			client.ProviderClient,
			viper.GetString("container"),
		)
		if err != nil {
			return fmt.Errorf("failed to initialize Swift container: %w", err)
		}

		// Use hyphenated timestamp format for safe and sortable filenames
		const timeFormat = "2006-01-02-150405"
		// Create and configure export file
		filename := viper.GetString("filename")
		if filename == "" {
			filename = "hermes-export-" + time.Now().Format(timeFormat)
		}

		format, err := parseExportFormat(viper.GetString("format"))
		if err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}
		exportFile := ExportFile{
			Format:      format,
			FileName:    filename,
			SegmentSize: uint64(viper.GetInt("segment-size")) * 1024 * 1024, // Convert MB to bytes
			Contents:    progressReader,
		}

		// Upload to Swift
		if err := exportFile.UploadTo(ctx, container); err != nil {
			return fmt.Errorf("failed to upload to Swift: %w", err)
		}

		fmt.Fprintf(os.Stderr, "\nSuccessfully exported %d events\n", len(allEvents))
		return nil
	},
}

// progressReader wraps an io.Reader to update a progress bar
type progressReader struct {
	Reader io.Reader
	Bar    *pb.ProgressBar
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		pr.Bar.Add(n)
	}
	return
}

func init() {
	initExportCmdFlags()
	RootCmd.AddCommand(ExportCmd)
}

func initExportCmdFlags() {
	ExportCmd.Flags().String("container", "", "Swift container name (required)")
	ExportCmd.Flags().String("format", "json", "Output format (json|csv|yaml)")
	ExportCmd.Flags().String("filename", "", "Name of the output file (default: hermes-export-{timestamp})")

	// Use same default as list command
	ExportCmd.Flags().UintP("limit", "l", maxOffset, "limit number of events to export (default: 10000)")

	// Hidden advanced options
	ExportCmd.Flags().Int("segment-size", 100, "Size of segments in MB for large file uploads")
	ExportCmd.Flags().MarkHidden("segment-size") //nolint:errcheck

	// Add all list command flags for filtering
	ExportCmd.Flags().StringP("target-type", "", "", "filter events by a target type")
	ExportCmd.Flags().StringP("target-id", "", "", "filter events by a target ID")
	ExportCmd.Flags().StringP("initiator-id", "", "", "filter events by an initiator ID")
	ExportCmd.Flags().StringP("initiator-name", "", "", "filter events by an initiator name")
	ExportCmd.Flags().StringP("action", "", "", "filter events by an action")
	ExportCmd.Flags().StringP("outcome", "", "", "filter events by an outcome")
	ExportCmd.Flags().StringP("time", "", "", "filter events by time")
	ExportCmd.Flags().StringP("time-start", "", "", "filter events from time")
	ExportCmd.Flags().StringP("time-end", "", "", "filter events till time")
	ExportCmd.Flags().StringP("project-id", "", "", "filter events by the project or domain ID (admin only)")
	ExportCmd.Flags().BoolP("all-projects", "A", false, "include all projects and domains (admin only)")
}

// buildListOpts creates ListOpts from viper flags
func buildListOpts() events.ListOpts {
	projectID := viper.GetString("project-id")
	if viper.GetBool("all-projects") {
		projectID = "*"
	}

	listOpts := events.ListOpts{
		Limit:         viper.GetInt("limit"),
		TargetType:    viper.GetString("target-type"),
		TargetID:      viper.GetString("target-id"),
		InitiatorID:   viper.GetString("initiator-id"),
		InitiatorName: viper.GetString("initiator-name"),
		Action:        viper.GetString("action"),
		Outcome:       viper.GetString("outcome"),
		ProjectID:     projectID,
	}

	// Handle time flags
	if t := viper.GetString("time"); t != "" {
		rt, err := parseTime(t)
		if err != nil {
			logg.Error("failed to parse time: %v", err)
			return listOpts
		}
		listOpts.Time = []events.DateQuery{{Date: rt}}
	}

	if t := viper.GetString("time-start"); t != "" {
		rt, err := parseTime(t)
		if err != nil {
			logg.Error("failed to parse time-start: %v", err)
			return listOpts
		}
		listOpts.Time = append(listOpts.Time, events.DateQuery{
			Date:   rt,
			Filter: events.DateFilterGTE,
		})
	}

	if t := viper.GetString("time-end"); t != "" {
		rt, err := parseTime(t)
		if err != nil {
			logg.Error("failed to parse time-end: %v", err)
			return listOpts
		}
		listOpts.Time = append(listOpts.Time, events.DateQuery{
			Date:   rt,
			Filter: events.DateFilterLTE,
		})
	}

	return listOpts
}
