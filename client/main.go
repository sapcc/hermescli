// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/utils/v2/client"
	"github.com/gophercloud/utils/v2/env"
	"github.com/gophercloud/utils/v2/openstack/clientconfig"
	"github.com/sapcc/go-bits/secrets"
	"github.com/sapcc/gophercloud-sapcc/v2/clients"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "hermescli",
	Short:        "Hermes CLI tool",
	SilenceUsage: true,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	initRootCmdFlags()
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initRootCmdFlags() {
	// debug flag
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "print out request and response objects")
	RootCmd.PersistentFlags().StringSliceP("column", "c", []string{}, "an event column to print")
	RootCmd.PersistentFlags().StringP("format", "f", "table", "the output format")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))   //nolint:errcheck
	viper.BindPFlag("column", RootCmd.PersistentFlags().Lookup("column")) //nolint:errcheck
	viper.BindPFlag("format", RootCmd.PersistentFlags().Lookup("format")) //nolint:errcheck
}

// NewHermesV1Client returns a *ServiceClient for making calls
// to the OpenStack Lyra v1 API. An error will be returned if
// authentication or client creation was not possible.
func NewHermesV1Client(ctx context.Context) (*gophercloud.ServiceClient, error) {
	if err := secrets.GetPasswordFromCommandIfRequested(); err != nil {
		return nil, err
	}
	ao, err := clientconfig.AuthOptions(nil)
	if err != nil {
		return nil, err
	}

	/* TODO: Introduce auth by CLI parameters
	   ao := gophercloud.AuthOptions{
	           IdentityEndpoint:            authURL,
	           UserID:                      userID,
	           Username:                    username,
	           Password:                    password,
	           TenantID:                    tenantID,
	           TenantName:                  tenantName,
	           DomainID:                    domainID,
	           DomainName:                  domainName,
	           ApplicationCredentialID:     applicationCredentialID,
	           ApplicationCredentialName:   applicationCredentialName,
	           ApplicationCredentialSecret: applicationCredentialSecret,
	   }
	*/

	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	if viper.GetBool("debug") {
		provider.HTTPClient = http.Client{
			Transport: &client.RoundTripper{
				Rt:     &http.Transport{},
				Logger: &client.DefaultLogger{},
			},
		}
	}

	err = openstack.Authenticate(ctx, provider, *ao)
	if err != nil {
		return nil, err
	}

	return clients.NewHermesV1(provider, gophercloud.EndpointOpts{
		Region: env.Getenv("OS_REGION_NAME"),
	})
}

func verifyGlobalFlags(columnsOrder []string) error {
	// verify supported columns
	columns := viper.GetStringSlice("column")
	for _, c := range columns {
		if len(columnsOrder) == 0 {
			return errors.New(`columns are not supported for this command`)
		}
		if !isSliceContainsStr(columnsOrder, c) {
			return fmt.Errorf(`invalid "%s" column name, supported values for the column: %s`, c, strings.Join(columnsOrder, ", "))
		}
	}

	// verify supported formats
	if !isSliceContainsStr(defaultPrintFormats, viper.GetString("format")) {
		return fmt.Errorf(`invalid "%s" column name, supported values for the format: %s`, viper.GetString("format"), strings.Join(defaultPrintFormats, ", "))
	}

	// verify the project ID and the domain ID parameters
	projectID := viper.GetString("project-id")
	allProjects := viper.GetBool("all-projects")
	if projectID != "" && allProjects {
		return errors.New("--project-id and --all-projects cannot be both specified")
	}

	return nil
}

// isSliceContainsStr returns true if the string exists in given slice
func isSliceContainsStr(sl []string, str string) bool {
	for _, s := range sl {
		if s == str {
			return true
		}
	}
	return false
}
