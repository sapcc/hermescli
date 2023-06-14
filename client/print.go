package client

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sapcc/gophercloud-sapcc/audit/v1/events"
	"gopkg.in/yaml.v2"
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

	if len(event.Observer.Name) > 0 {
		kv["Observer"] = event.Observer.Name
	}
	kv["TypeURI"] = event.Observer.TypeURI
	// compatibility to Source<->Observer.TypeURI link
	kv["Source"] = event.Observer.TypeURI

	kv["Action"] = string(event.Action)
	kv["Outcome"] = string(event.Outcome)
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

	if len(event.Initiator.AppCredentialID) > 0 {
		kv["InitiatorAppCredential"] = event.Initiator.AppCredentialID
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
		return fmt.Errorf("error writing header to csv: %s", err)
	}

	for _, v := range allEvents {
		kv := eventToKV(v)
		tableRow := []string{}
		for _, k := range keyOrder {
			v := kv[k]
			tableRow = append(tableRow, v)
		}
		if err := csv.Write(tableRow); err != nil {
			return fmt.Errorf("error writing record to csv: %s", err)
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
