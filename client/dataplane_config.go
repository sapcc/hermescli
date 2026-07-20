// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/dataplaneconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DataplaneConfigCmd is the parent command for dataplane-config subcommands.
var DataplaneConfigCmd = &cobra.Command{
	Use:   "dataplane-config",
	Short: "Manage Hermes dataplane event configuration",
}

var dataplaneConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get dataplane event configuration for a project",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := resolveProjectID()
		if err != nil {
			return err
		}

		client, err := NewHermesV1Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		cfg, err := dataplaneconfig.Get(cmd.Context(), client, projectID).Extract()
		if err != nil {
			return err
		}

		return printDataplaneConfig(cfg)
	},
}

var dataplaneConfigSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or replace dataplane event configuration for a project",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := resolveProjectID()
		if err != nil {
			return err
		}

		client, err := NewHermesV1Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		opts := dataplaneconfig.PutOpts{
			Enabled:      viper.GetBool("enabled"),
			TargetBucket: viper.GetString("target-bucket"),
		}

		cfg, err := dataplaneconfig.Put(cmd.Context(), client, projectID, opts).Extract()
		if err != nil {
			return err
		}

		return printDataplaneConfig(cfg)
	},
}

var dataplaneConfigDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete dataplane event configuration for a project",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := resolveProjectID()
		if err != nil {
			return err
		}

		client, err := NewHermesV1Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		if err := dataplaneconfig.Delete(cmd.Context(), client, projectID).ExtractErr(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "dataplane-config for project %s deleted\n", projectID)
		return nil
	},
}

func init() {
	DataplaneConfigCmd.AddCommand(dataplaneConfigGetCmd)
	DataplaneConfigCmd.AddCommand(dataplaneConfigSetCmd)
	DataplaneConfigCmd.AddCommand(dataplaneConfigDeleteCmd)
	RootCmd.AddCommand(DataplaneConfigCmd)

	dataplaneConfigGetCmd.Flags().String("project-id", "", "project ID (defaults to OS_PROJECT_ID)")
	dataplaneConfigSetCmd.Flags().String("project-id", "", "project ID (defaults to OS_PROJECT_ID)")
	dataplaneConfigSetCmd.Flags().Bool("enabled", false, "enable dataplane event routing")
	dataplaneConfigSetCmd.Flags().String("target-bucket", "", "target S3 bucket name")
	dataplaneConfigDeleteCmd.Flags().String("project-id", "", "project ID (defaults to OS_PROJECT_ID)")
}

// resolveProjectID returns the project ID from --project-id flag or OS_PROJECT_ID env var.
func resolveProjectID() (string, error) {
	if id := viper.GetString("project-id"); id != "" {
		return id, nil
	}
	if id := os.Getenv("OS_PROJECT_ID"); id != "" {
		return id, nil
	}
	return "", errors.New("--project-id is required (or set OS_PROJECT_ID)")
}

func printDataplaneConfig(cfg *dataplaneconfig.DataplaneConfig) error {
	switch viper.GetString("format") {
	case "json":
		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", out)
	case "yaml":
		out, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Printf("%s", out)
	default: // table, value
		fmt.Printf("%-15s %s\n", "ProjectID", cfg.ProjectID)
		fmt.Printf("%-15s %v\n", "Enabled", cfg.Enabled)
		fmt.Printf("%-15s %s\n", "TargetBucket", cfg.TargetBucket)
		fmt.Printf("%-15s %s\n", "UpdatedAt", cfg.UpdatedAt)
		fmt.Printf("%-15s %s\n", "UpdatedBy", cfg.UpdatedBy)
	}
	return nil
}
