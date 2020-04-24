package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/sapcc/gophercloud-sapcc/audit/v1/attributes"
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
	Use:       fmt.Sprintf("attributes %s", strings.Join(validArgs, "|")),
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
		client, err := NewHermesV1Client()
		if err != nil {
			return fmt.Errorf("Failed to create Hermes client: %s", err)
		}

		format := viper.GetString("format")

		listOpts := attributes.ListOpts{
			Limit:     viper.GetInt("limit"),
			MaxDepth:  viper.GetInt("max-depth"),
			ProjectID: viper.GetString("project-id"),
			DomainID:  viper.GetString("domain-id"),
		}

		var allAttributes []string

		for _, name := range args {
			err = attributes.List(client, name, listOpts).EachPage(func(page pagination.Page) (bool, error) {
				attrs, err := attributes.ExtractAttributes(page)
				if err != nil {
					return false, fmt.Errorf("Failed to extract attributes: %s", err)
				}

				allAttributes = append(allAttributes, attrs...)

				return true, nil
			})
			if err != nil {
				if _, ok := err.(gophercloud.ErrDefault500); ok {
					return fmt.Errorf(`Failed to list attributes: %s: please try to decrease an amount of the attributes in output, e.g. set "--limit 100"`, err)
				}
				return fmt.Errorf("Failed to list attributes: %s", err)
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
			return fmt.Errorf("Unsupported format: %s", format)
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
	AttributesCmd.Flags().StringP("project-id", "", "", "filter attributes by the project ID")
	AttributesCmd.Flags().StringP("domain-id", "", "", "filter attributes by the domain ID")
}
