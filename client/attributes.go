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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/attributes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var validArgs = []string{
	"observer_type",
	"target_type",
	"target_id",
	"initiator_type",
	"initiator_id",
	"initiator_name",
	"action",
	"outcome"}

func validateAll(checks ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, check := range checks {
			if err := check(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// AttributesCmd represents the list command
var AttributesCmd = &cobra.Command{
	Use:       "attributes " + strings.Join(validArgs, "|"),
	Args:      validateAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: validArgs,
	Short:     "List Hermes attributes",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		return verifyGlobalFlags(nil)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// list attributes
		client, err := NewHermesV1Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create Hermes client: %w", err)
		}

		format := viper.GetString("format")

		projectID := viper.GetString("project-id")
		if viper.GetBool("all-projects") {
			projectID = "*"
		}

		listOpts := attributes.ListOpts{
			Limit:     viper.GetInt("limit"),
			MaxDepth:  viper.GetInt("max-depth"),
			ProjectID: projectID,
		}

		var allAttributes []string

		for _, name := range args {
			err = attributes.List(client, name, listOpts).EachPage(cmd.Context(), func(ctx context.Context, page pagination.Page) (bool, error) {
				attrs, err := attributes.ExtractAttributes(page)
				if err != nil {
					return false, fmt.Errorf("failed to extract attributes: %w", err)
				}

				allAttributes = append(allAttributes, attrs...)

				return true, nil
			})
			if err != nil {
				if gophercloud.ResponseCodeIs(err, http.StatusInternalServerError) {
					return fmt.Errorf(`failed to list attributes: %w: please try to decrease the amount of the attributes in output, e.g. set "--limit 100"`, err)
				}
				return fmt.Errorf("failed to list attributes: %w", err)
			}
		}

		switch format {
		case "json":
			jsonAttrs, err := json.MarshalIndent(allAttributes, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", jsonAttrs)
		case "yaml":
			yamlAttrs, err := yaml.Marshal(allAttributes)
			if err != nil {
				return err
			}
			fmt.Printf("%s", yamlAttrs)
		case "csv", "value", "table":
			fmt.Printf("%s\n", strings.Join(allAttributes, "\n"))
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}

		return nil
	},
}

func init() {
	initAttributesCmdFlags()
	RootCmd.AddCommand(AttributesCmd)
}

func initAttributesCmdFlags() {
	AttributesCmd.Flags().UintP("limit", "l", 0, "limit an amount of attributes in output")
	AttributesCmd.Flags().UintP("max-depth", "", 0, "limit the level of detail of hierarchical values")
	AttributesCmd.Flags().StringP("project-id", "", "", "filter attributes by the project or domain ID (admin only)")
	AttributesCmd.Flags().BoolP("all-projects", "A", false, "include all projects and domains (admin only) (alias for --project-id '*')")
}
