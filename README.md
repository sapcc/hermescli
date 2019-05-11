# gophercloud-hermes

# Usage

Only environment authentication metod is supported.

## List

### Usage

```sh
List Hermes events

Usage:
  hermesctl list [flags]

Flags:
      --action string           filter events by an action
  -h, --help                    help for list
      --initiator-name string   filter events by an initiator name
      --outcome string          filter events by an outcome
      --sort strings            Supported sort keys include time, observer_type, target_type, target_id, initiator_type, initiator_id, outcome and action.
                                Each sort key may also include a direction suffix.
                                Supported directions are :asc for ascending and :desc for descending.
                                Can be specified multiple times.
      --source string           filter events by a source
      --target-type string      filter events by a target type
      --time string             filter events by time
      --time-end string         filter events till time
      --time-start string       filter events from time

Global Flags:
      --debug   Print out request and response objects.
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
  hermesctl show [flags] <event-id>

Flags:
  -h, --help   help for show

Global Flags:
      --debug   Print out request and response objects.
```

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
