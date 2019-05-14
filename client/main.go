package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/kayrus/gophercloud-hermes/audit"
	"github.com/kayrus/gophercloud-hermes/audit/v1/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultPrintFormats = []string{
	"table",
	"value",
	"json",
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "hermesctl",
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
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("column", RootCmd.PersistentFlags().Lookup("column"))
	viper.BindPFlag("format", RootCmd.PersistentFlags().Lookup("format"))
}

func main() {
	// Workaround for the "AuthOptionsFromEnv"
	os.Setenv("OS_DOMAIN_NAME", os.Getenv("OS_PROJECT_DOMAIN_NAME"))

	Execute()
}

// NewHermesV1Client returns a *ServiceClient for making calls
// to the OpenStack Lyra v1 API. An error will be returned if
// authentication or client creation was not possible.
func NewHermesV1Client() (*gophercloud.ServiceClient, error) {
	ao, err := openstack.AuthOptionsFromEnv()
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

	client, err := openstack.AuthenticatedClient(ao)
	if err != nil {
		return nil, err
	}

	if viper.GetBool("debug") {
		client.HTTPClient = http.Client{
			Transport: &clients.LogRoundTripper{
				Rt: &http.Transport{},
			},
		}
	}
	return audit.NewHermesV1(client, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}

func eventToKV(event events.Event) map[string]string {
	kv := make(map[string]string)
	kv["ID"] = event.ID
	kv["Type"] = event.EventType
	kv["Time"] = event.EventTime

	if len(event.Observer.Name) > 0 {
		kv["Observer"] = event.Observer.Name
	}
	kv["TypeURI"] = event.Observer.TypeURI
	// compatibility to Source<->Observer.TypeURI link
	kv["Source"] = event.Observer.TypeURI

	kv["Action"] = event.Action
	kv["Outcome"] = event.Outcome
	kv["Target"] = fmt.Sprintf("%s %s", event.Target.TypeURI, event.Target.ID)

	if len(event.Initiator.Name) > 0 {
		kv["Initiator"] = event.Initiator.Name
	}
	if len(event.Initiator.Domain) > 0 {
		kv["InitiatorDomain"] = event.Initiator.Domain
	}
	if event.Initiator.Host != nil {
		kv["InitiatorAddress"] = event.Initiator.Host.Address
		kv["InitiatorAgent"] = event.Initiator.Host.Agent
	}

	if len(event.RequestPath) > 0 {
		kv["RequestPath"] = event.RequestPath
	}

	if len(event.Attachments) > 0 {
		var attachments []string
		for _, attachment := range event.Attachments {
			if attachment.Content != nil {
				attachments = append(attachments, fmt.Sprintf("%v", attachment.Content))
			}
		}
		if len(attachments) > 0 {
			kv["Attachments"] = strings.Join(attachments, "\n")
		}
	}
	return kv
}

func verifyGlobalFlags(columnsOrder []string) error {
	// verify supported columns
	columns := viper.GetStringSlice("column")
	for _, c := range columns {
		if !isSliceContainsStr(columnsOrder, c) {
			return fmt.Errorf(`Invalid "%s" column name, supported values for the column: %s`, c, strings.Join(columnsOrder, ", "))
		}
	}

	// verify supported formats
	if !isSliceContainsStr(defaultPrintFormats, viper.GetString("format")) {
		return fmt.Errorf(`Invalid "%s" column name, supported values for the format: %s`, viper.GetString("format"), strings.Join(defaultPrintFormats, ", "))
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
