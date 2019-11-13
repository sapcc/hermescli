# hermes-ctl

# Usage

Only environment authentication method is supported.

## List

### Usage

```sh
List Hermes events

Usage:
  hermesctl list [flags]

Flags:
      --action string           filter events by an action
      --domain-id string        filter events by the domain ID
  -h, --help                    help for list
      --initiator-id string     filter events by an initiator ID
      --initiator-name string   filter events by an initiator name
  -l, --limit uint              limit an amount of events in output
      --outcome string          filter events by an outcome
      --over-10k-fix            workaround to filter out overlapping events for > 10k total events (default true)
      --project-id string       filter events by the project ID
  -s, --sort strings            supported sort keys include time, observer_type, target_type, target_id, initiator_type, initiator_id, outcome and action
                                each sort key may also include a direction suffix
                                supported directions are ":asc" for ascending and ":desc" for descending
                                can be specified multiple times
      --source string           filter events by a source
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
$ hermesctl list --time 2019-04-23T22:07:16+0000 --sort time:asc
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
  hermesctl show <event-id> [<event-id>...] [flags]

Flags:
      --domain-id string    show event for the domain ID
  -h, --help                help for show
      --project-id string   show event for the project ID

Global Flags:
  -c, --column strings   an event column to print
  -d, --debug            print out request and response objects
  -f, --format string    the output format (default "table")
```

### Example

```sh
$ hermesctl show 1878df7c-d3ec-52d0-8b56-11ad68d25102
+------------------+--------------------------------------------------+
|       KEY        |                      VALUE                       |
+------------------+--------------------------------------------------+
| ID               | 1878df7c-d3ec-52d0-8b56-11ad68d25102             |
| Type             | activity                                         |
| Time             | 2019-04-23T22:07:16+0000                         |
| Observer         | neutron                                          |
| TypeURI          | service/network                                  |
| Action           | update                                           |
| Outcome          | success                                          |
| Target           | network/port                                     |
|                  | 88c4c917-f5de-43e5-a403-b7c023bfc13d             |
| Initiator        | neutron                                          |
| InitiatorDomain  | Default                                          |
| InitiatorAddress | 100.65.0.80                                      |
| InitiatorAgent   | python-neutronclient                             |
| RequestPath      | /v2.0/ports/88c4c917-f5de-43e5-a403-b7c023bfc13d |
+------------------+--------------------------------------------------+
```

## Attributes

### Usage

```sh
List Hermes attributes

Usage:
  hermesctl attributes observer_type|target_type|target_id|initiator_type|initiator_id|initiator_name|action|outcome [flags]

Flags:
      --domain-id string    filter attributes by the domain ID
  -h, --help                help for attributes
  -l, --limit uint          limit an amount of attributes in output
      --max-depth uint      limit the level of detail of hierarchical values
      --project-id string   filter attributes by the project ID

Global Flags:
  -c, --column strings   an event column to print
  -d, --debug            print out request and response objects
  -f, --format string    the output format (default "table")
```

### Example

```sh
$ hermesctl attributes outcome
success
failure
unknown
```

## Build

```sh
$ make
# or within the docker container
$ make docker
```
