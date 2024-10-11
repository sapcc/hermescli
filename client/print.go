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
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sapcc/gophercloud-sapcc/v2/audit/v1/events"
	"gopkg.in/yaml.v3"
)

var defaultPrintFormats = []string{
	"table",
	"value",
	"json",
	"csv",
	"yaml",
}

func eventToKV(event events.Event) map[string]string {
	kv := make(map[string]string)
	kv["ID"] = event.ID
	kv["Type"] = event.EventType
	kv["Time"] = event.EventTime

	if event.Observer.Name != "" {
		kv["Observer"] = event.Observer.Name
	}
	kv["TypeURI"] = event.Observer.TypeURI
	// compatibility to Source<->Observer.TypeURI link
	kv["Source"] = event.Observer.TypeURI

	kv["Action"] = string(event.Action)
	kv["Outcome"] = string(event.Outcome)
	kv["Target"] = fmt.Sprintf("%s %s", event.Target.TypeURI, event.Target.ID)

	if event.Initiator.Name != "" {
		kv["Initiator"] = event.Initiator.Name
	}
	if event.Initiator.Domain != "" {
		kv["InitiatorDomain"] = event.Initiator.Domain
	}
	if event.Initiator.Host != nil {
		kv["InitiatorAddress"] = event.Initiator.Host.Address
		kv["InitiatorAgent"] = event.Initiator.Host.Agent
	}

	if event.Initiator.AppCredentialID != "" {
		kv["InitiatorAppCredential"] = event.Initiator.AppCredentialID
	}

	if event.RequestPath != "" {
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

func printEvent(allEvents []events.Event, format string, keyOrder []string) error {
	switch format {
	case "json":
		return printJSON(allEvents)
	case "yaml":
		return printYAML(allEvents)
	case "csv":
		return printCSV(allEvents, keyOrder)
	case "value":
		return printValue(allEvents, keyOrder)
	}
	return fmt.Errorf("unsupported format: %s", format)
}

func printCSV(allEvents []events.Event, keyOrder []string) error {
	var buf bytes.Buffer
	csv := csv.NewWriter(&buf) //nolint:gocritic

	if err := csv.Write(keyOrder); err != nil {
		return fmt.Errorf("error writing header to csv: %w", err)
	}

	for _, v := range allEvents {
		kv := eventToKV(v)
		tableRow := []string{}
		for _, k := range keyOrder {
			v := kv[k]
			tableRow = append(tableRow, v)
		}
		if err := csv.Write(tableRow); err != nil {
			return fmt.Errorf("error writing record to csv: %w", err)
		}
	}

	csv.Flush()

	fmt.Print(buf.String())

	return nil
}

func printJSON(allEvents []events.Event) error {
	if len(allEvents) > 1 {
		jsonEvents, err := json.MarshalIndent(allEvents, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonEvents)
	} else if len(allEvents) == 1 {
		jsonEvent, err := json.MarshalIndent(allEvents[0], "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonEvent)
	}
	return nil
}

func printYAML(allEvents []events.Event) error {
	if len(allEvents) > 1 {
		yamlEvents, err := yaml.Marshal(allEvents)
		if err != nil {
			return err
		}
		fmt.Printf("%s", yamlEvents)
	} else if len(allEvents) == 1 {
		yamlEvent, err := yaml.Marshal(allEvents[0])
		if err != nil {
			return err
		}
		fmt.Printf("%s", yamlEvent)
	}
	return nil
}

func printValue(allEvents []events.Event, keyOrder []string) error {
	for _, v := range allEvents {
		kv := eventToKV(v)
		var p []string
		for _, k := range keyOrder {
			v := kv[k]
			p = append(p, v)
		}
		fmt.Printf("%s\n", strings.Join(p, " "))
	}
	return nil
}
