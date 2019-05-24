package client

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/sapcc/hermes-ctl/audit"
	"github.com/sapcc/hermes-ctl/audit/v1/events"
	"github.com/sapcc/hermes-ctl/env"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultPrintFormats = []string{
	"table",
	"value",
	"json",
	"csv",
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

// NewHermesV1Client returns a *ServiceClient for making calls
// to the OpenStack Lyra v1 API. An error will be returned if
// authentication or client creation was not possible.
func NewHermesV1Client() (*gophercloud.ServiceClient, error) {
	ao, err := AuthOptionsFromEnv()
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

	client, err := openstack.NewClient(ao.IdentityEndpoint)
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

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return nil, err
	}

	return audit.NewHermesV1(client, gophercloud.EndpointOpts{
		Region: env.Get("OS_REGION_NAME"),
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

	var attachments []string
	for _, attachment := range event.Attachments {
		if attachment.Content != nil {
			attachments = append(attachments, fmt.Sprintf("%v", attachment.Content))
		}
	}
	for _, attachment := range event.Target.Attachments {
		if attachment.Content != nil {
			attachments = append(attachments, fmt.Sprintf("%v", attachment.Content))
		}
	}
	if len(attachments) > 0 {
		kv["Attachments"] = strings.Join(attachments, "\n")
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

var nilOptions = gophercloud.AuthOptions{}

/*
AuthOptionsFromEnv fills out an identity.AuthOptions structure with the
settings found on the various OpenStack OS_* environment variables.

The following variables provide sources of truth: OS_AUTH_URL, OS_USERNAME,
OS_PASSWORD and OS_PROJECT_ID.

Of these, OS_USERNAME, OS_PASSWORD, and OS_AUTH_URL must have settings,
or an error will result.  OS_PROJECT_ID, is optional.

OS_TENANT_ID and OS_TENANT_NAME are deprecated forms of OS_PROJECT_ID and
OS_PROJECT_NAME and the latter are expected against a v3 auth api.

If OS_PROJECT_ID and OS_PROJECT_NAME are set, they will still be referred
as "tenant" in Gophercloud.

If OS_PROJECT_NAME is set, it requires OS_PROJECT_ID to be set as well to
handle projects not on the default domain.

To use this function, first set the OS_* environment variables (for example,
by sourcing an `openrc` file), then:

	opts, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClient(opts)
*/
func AuthOptionsFromEnv() (gophercloud.AuthOptions, error) {
	authURL := env.Get("OS_AUTH_URL")
	username := env.Get("OS_USERNAME")
	userID := env.Get("OS_USERID")
	password := env.Get("OS_PASSWORD")
	tenantID := env.Get("OS_TENANT_ID")
	tenantName := env.Get("OS_TENANT_NAME")
	domainID := env.Get("OS_DOMAIN_ID")
	domainName := env.Get("OS_DOMAIN_NAME")
	applicationCredentialID := env.Get("OS_APPLICATION_CREDENTIAL_ID")
	applicationCredentialName := env.Get("OS_APPLICATION_CREDENTIAL_NAME")
	applicationCredentialSecret := env.Get("OS_APPLICATION_CREDENTIAL_SECRET")

	token := env.Get("OS_AUTH_TOKEN")
	if token == "" {
		// fallback to an old env name
		token = env.Get("OS_TOKEN")
	}

	// If OS_PROJECT_ID is set, overwrite tenantID with the value.
	if v := env.Get("OS_PROJECT_ID"); v != "" {
		tenantID = v
	}

	// If OS_PROJECT_NAME is set, overwrite tenantName with the value.
	if v := env.Get("OS_PROJECT_NAME"); v != "" {
		tenantName = v
	}

	// If OS_PROJECT_DOMAIN_NAME is set, overwrite domainName with the value.
	if v := env.Get("OS_PROJECT_DOMAIN_NAME"); v != "" {
		domainName = v
	}

	// If OS_PROJECT_DOMAIN_ID is set, overwrite domainID with the value.
	if v := env.Get("OS_PROJECT_DOMAIN_ID"); v != "" {
		domainID = v
	}

	if authURL == "" {
		err := gophercloud.ErrMissingEnvironmentVariable{
			EnvironmentVariable: "OS_AUTH_URL",
		}
		return nilOptions, err
	}

	if userID == "" && username == "" && token == "" {
		// Empty username and userID could be ignored, when applicationCredentialID and applicationCredentialSecret are set
		if applicationCredentialID == "" && applicationCredentialSecret == "" {
			err := gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
				EnvironmentVariables: []string{"OS_USERID", "OS_USERNAME", "OS_AUTH_TOKEN"},
			}
			return nilOptions, err
		}
	}

	if password == "" && applicationCredentialID == "" && applicationCredentialName == "" && token == "" {
		err := gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
			EnvironmentVariables: []string{"OS_PASSWORD", "OS_AUTH_TOKEN"},
		}
		return nilOptions, err
	}

	if (applicationCredentialID != "" || applicationCredentialName != "") && applicationCredentialSecret == "" {
		err := gophercloud.ErrMissingEnvironmentVariable{
			EnvironmentVariable: "OS_APPLICATION_CREDENTIAL_SECRET",
		}
		return nilOptions, err
	}

	if domainID == "" && domainName == "" && tenantID == "" && tenantName != "" {
		err := gophercloud.ErrMissingEnvironmentVariable{
			EnvironmentVariable: "OS_PROJECT_ID",
		}
		return nilOptions, err
	}

	if applicationCredentialID == "" && applicationCredentialName != "" && applicationCredentialSecret != "" {
		if userID == "" && username == "" && token == "" {
			return nilOptions, gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
				EnvironmentVariables: []string{"OS_USERID", "OS_USERNAME", "OS_AUTH_TOKEN"},
			}
		}
		if username != "" && domainID == "" && domainName == "" {
			return nilOptions, gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
				EnvironmentVariables: []string{"OS_DOMAIN_ID", "OS_DOMAIN_NAME"},
			}
		}
	}

	var scope *gophercloud.AuthScope
	if token != "" {
		// scope is required for the token auth
		username = ""
		userID = ""
		password = ""

		scope = &gophercloud.AuthScope{
			ProjectID:   tenantID,
			ProjectName: tenantName,
			DomainID:    domainID,
			DomainName:  domainName,
		}

		domainName = ""
		domainID = ""

		tenantName = ""
		tenantID = ""
	}

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
		TokenID:                     token,
		Scope:                       scope,
	}

	return ao, nil
}

func printCSV(allEvents []events.Event, keyOrder []string) error {
	var buf bytes.Buffer
	csv := csv.NewWriter(&buf)

	if err := csv.Write(keyOrder); err != nil {
		return fmt.Errorf("error writing header to csv:", err)
	}

	for _, v := range allEvents {
		kv := eventToKV(v)
		tableRow := []string{}
		for _, k := range keyOrder {
			v, _ := kv[k]
			tableRow = append(tableRow, v)
		}
		if err := csv.Write(tableRow); err != nil {
			return fmt.Errorf("error writing record to csv:", err)
		}
	}

	csv.Flush()

	fmt.Print(buf.String())

	return nil
}
