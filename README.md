# Hermescli

[![GitHub Release](https://img.shields.io/github/v/release/sapcc/hermescli)](https://github.com/sapcc/hermescli/releases/latest)
[![CI](https://github.com/sapcc/hermescli/actions/workflows/ci.yaml/badge.svg)](https://github.com/sapcc/hermescli/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sapcc/hermescli)](https://goreportcard.com/report/github.com/sapcc/hermescli)

Hermes CLI is a command line interface for interacting with [Hermes](https://github.com/sapcc/hermes), an OpenStack service for storing and retrieving audit events. It allows users to easily retrieve and display audit events from the service, without the need for manual API calls or a separate client library.

## Installation

We provide pre-compiled binaries for the [latest release](https://github.com/sapcc/hermescli/releases/latest).

Alternatively, you can build with `make` or install with `make install`. The latter
understands the conventional environment variables for choosing install locations:
`DESTDIR` and `PREFIX`.

## Commands

Hermes CLI offers the following commands:

- `list`: Retrieve a list of audit events
- `show`: Show details for a specific event
- `attributes`: List attributes related to audit events
- `export`: Export events to Swift

## Usage

```sh
List Hermes events

Usage:
  hermescli list [flags]

Flags:
      --action string           filter events by an action
  -A, --all-projects            include all projects and domains (admin only) (alias for --project-id '*')
  -h, --help                    help for list
      --initiator-id string     filter events by an initiator ID
      --initiator-name string   filter events by an initiator name
  -l, --limit uint              limit an amount of events in output
      --outcome string          filter events by an outcome
      --over-10k-fix            workaround to filter out overlapping events for > 10k total events (default true)
      --project-id string       filter events by the project or domain ID (admin only)
  -s, --sort strings            supported sort keys include time, observer_type, target_type, target_id, initiator_type, initiator_id, outcome and action
                                each sort key may also include a direction suffix
                                supported directions are ":asc" for ascending and ":desc" for descending
                                can be specified multiple times
      --source string           filter events by a source
      --search string           filter events by a full event search
      --target-id string        filter events by a target ID
      --target-type string      filter events by a target type
      --time string             filter events by time
      --time-end string         filter events till time
      --time-start string       filter events from time

Global Flags:
  -c, --column strings   an event column to print
  -d, --debug            print out request and response objects
  -f, --format string    the output format (default "table")
```

### Example

```sh
$ hermescli list --time 2019-04-23T22:07:16+0000 --sort time:asc
+--------------------------------------+--------------------------+-----------------+--------+---------+--------------------------------------+-----------+
|                  ID                  |           TIME           |     SOURCE      | ACTION | OUTCOME |                TARGET                | INITIATOR |
+--------------------------------------+--------------------------+-----------------+--------+---------+--------------------------------------+-----------+
| 1878df7c-d3ec-52d0-8b56-11ad68d25102 | 2019-04-23T22:07:16+0000 | service/network | update | success | network/port                         | neutron   |
|                                      |                          |                 |        |         | 88c4c917-f5de-43e5-a403-b7c023bfc13d |           |
+--------------------------------------+--------------------------+-----------------+--------+---------+--------------------------------------+-----------+
```

## Show

### Usage

```sh
Show Hermes event

Usage:
  hermescli show <event-id> [<event-id>...] [flags]

Flags:
  -A, --all-projects        include all projects and domains (admin only) (alias for --project-id '*')
  -h, --help                help for show
      --project-id string   show event for the project or domain ID (admin only)

Global Flags:
  -c, --column strings   an event column to print
  -d, --debug            print out request and response objects
  -f, --format string    the output format (default "table")
```

### Example

```sh
$ hermescli show 1878df7c-d3ec-52d0-8b56-11ad68d25102
+-------------------------+--------------------------------------------------+
|       KEY               |                      VALUE                       |
+-------------------------+--------------------------------------------------+
| ID                      | 1878df7c-d3ec-52d0-8b56-11ad68d25102             |
| Type                    | activity                                         |
| Time                    | 2019-04-23T22:07:16+0000                         |
| Observer                | neutron                                          |
| TypeURI                 | service/network                                  |
| Action                  | update                                           |
| Outcome                 | success                                          |
| Target                  | network/port                                     |
|                         | 88c4c917-f5de-43e5-a403-b7c023bfc13d             |
| Initiator               | neutron                                          |
| InitiatorDomain         | Default                                          |
| InitiatorAddress        | 100.65.0.80                                      |
| InitiatorAgent          | python-neutronclient                             |
| InitiatorAppCredential  | ee1246022693405e81b4e12fac1111cd                 |
| RequestPath             | /v2.0/ports/88c4c917-f5de-43e5-a403-b7c023bfc13d |
+-------------------------+--------------------------------------------------+
```

## Attributes

### Usage

`hermescli` requires the full set of OpenStack auth environment
variables. See [documentation for openstackclient](https://docs.openstack.org/python-openstackclient/latest/cli/man/openstack.html) for details.

```sh
List Hermes attributes

Usage:
  hermescli attributes observer_type|target_type|target_id|initiator_type|initiator_id|initiator_name|action|outcome [flags]

Flags:
  -A, --all-projects        include all projects and domains (admin only) (alias for --project-id '*')
  -h, --help                help for attributes
  -l, --limit uint          limit an amount of attributes in output
      --max-depth uint      limit the level of detail of hierarchical values
      --project-id string   filter attributes by the project or domain ID (admin only)

Global Flags:
  -c, --column strings   an event column to print
  -d, --debug            print out request and response objects
  -f, --format string    the output format (default "table")
```

### Example

```sh
$ hermescli attributes outcome
success
failure
unknown
```

## Export

### Usage

```sh
Export audit events to Swift storage

Usage:
  hermescli export [flags]

Flags:
      --container string       Swift container name (required)
      --format string         Output format (json|csv|yaml) (default "json")
      --filename string       Name of the output file (default "hermes-export-{timestamp}")
  -l, --limit uint           limit number of events to export (default: 10000)
      --time string          filter events by time
      --time-start string    filter events from time
      --time-end string      filter events till time
      --action string        filter events by action
      --outcome string       filter events by outcome
      --target-id string     filter events by a target ID
      --target-type string   filter events by a target type
      --initiator-id string  filter events by an initiator ID
      --initiator-name string filter events by an initiator name
      --project-id string    filter events by the project or domain ID (admin only)
  -A, --all-projects         include all projects and domains (admin only)

Global Flags:
  -d, --debug            print out request and response objects
  -f, --format string    the output format (default "table")
```

### Examples

```sh
# Export last week's events to JSON
$ hermescli export --container audit-exports --time-start "2024-01-07T00:00:00" --time-end "2024-01-14T23:59:59"
Fetching events...
Found 857 events to export
Converting to json format...
Uploading 2.3MB to Swift...
[==================================] 2.3MB/2.3MB
Successfully exported 857 events

# Export specific events as CSV
$ hermescli export --container audit-exports --format csv --initiator-name admin --action update
Fetching events...
Found 124 events to export
Converting to csv format...
Uploading 0.5MB to Swift...
[==================================] 0.5MB/0.5MB
Successfully exported 124 events
```

The `export` command allows you to export audit events to Swift storage for archival or further processing. Events can be exported in JSON, CSV, or YAML formats. The command supports all filtering options available in the `list` command.

By default, it will export up to 10,000 events. Use the `--limit` flag to adjust this number. Large exports are automatically handled through Swift's segmented upload feature.

> Note: This command requires Swift storage access in addition to the standard OpenStack authentication environment variables.

## Build

```sh
$ make
# or within the docker container
$ make docker
```

## Contributions

We welcome contributions to the Hermes CLI in the form of bug reports, feature requests, and pull requests.
