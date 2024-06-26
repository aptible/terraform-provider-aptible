# Aptible Log Drain Resource

This resource is used to create and manage
[Log Drains](https://deploy-docs.aptible.com/docs/log-drains)
running on Aptible Deploy.

## Example Usage

```hcl
resource "aptible_log_drain" "syslog_log_drain" {
    env_id = 123
    handle = "syslog_log_drain"
    drain_type = "syslog_tls_tcp"
    drain_host = "syslog.aptible.com"
    drain_port = "1234"
}
```

```hcl
resource "aptible_log_drain" "http_log_drain" {
    env_id = 123
    handle = "http_log_drain"
    drain_type = "https_post"
    url = "https://test.aptible.com"
    drain_apps = false
    drain_proxies = true
}
```

## Argument Reference

The required arguments vary based on `drain_type`.

Because we support a lot of different destinations for log drains, we could not
properly validate all the required fields for each `drain_type` within
terraform.

It's entirely possible that terraform will pass its linting but the log drain
isn't necessarily set up properly in Aptible.

- `env_id` (Required) - The ID of the environment you would like to create your
  Log Drain in. See main provider documentation for more on how to determine
  what you should use for `env_id`.
- `handle` (Required) - The handle for the log drain. This must be all lower
  case, and only contain letters, numbers, `-`, `_`, or `.`.
- `drain_type` (Required) - The type of log drain: `syslog_tls_tcp`,
  `elasticsearch_database`, `https_post`, `sumologic`, `logdna`, `datadog`,
  `papertrail`.
- `drain_apps` - (Optional, default `True`) If the drain should collect logs
  from apps.
- `drain_databases` - (Optional, default `True`) If the drain should collect
  logs from databases.
- `drain_proxies` - (Optional, default `False`) If the drain should collect logs
  from Endpoints.
- `drain_ephemeral_sessions` - (Optional, default `False`) If the drain should
  collect logs from SSH sessions.
- `database_id` - The ID of the elasticsearch database that
  `elasticsearch_database` drains should send logs to.
- `drain_host` - The host name of the destination to send logs to.
- `drain_port` - The port for the destination where logs drains will be sent.
- `logging_token` - The logging token prepended to logs by `syslog` and
  `papertrail` drains.
- `url` - The destination url where the logs will be sent.
- `tags` - A comma-separated list of additional tags to apply logs collected by
  `logdna` and `datadog` drains.
- `token` - The API token used by `logdna` and `datadog` drains.
- `pipeline` - The ID of the elasticsearch
  [ingest pipeline](https://www.elastic.co/guide/en/elasticsearch/reference/7.10/ingest.html)
  to use with `elasticsearch_database` drains.

### Arguments based on `drain_type`

Note that using additional, unsupported arguments for the given `drain_type` may
result in unnecessary resource replacement.

The following arguments are required for all log drains:

- `env_id`
- `handle`
- `drain_type`

#### `syslog_tls_tcp`

- `drain_host` (Required)
- `drain_port` (Required)
- `logging_token` (Optional)

#### `elasticsearch_database`

- `database_id` (Required)
- `pipeline` (Optional)

#### `https_post`

- `url` (Required)

#### `sumologic`

- `url` (Required)

#### `logdna`

- `token` (Required)
- `drain_host` (Optional)
- `tags` (Optional)

#### `datadog`

- `token` (Required)
- `drain_host` (Optional)
- `tags` (Optional)

#### `papertrail`

- `drain_host` (Required)
- `drain_port` (Required)
- `logging_token` (Optional)

## Attribute Reference

In addition to all the arguments listed above, the following attributes are 
exported:

- `log_drain_id` - The unique ID for the log drain.

## Import

Existing log drains can be imported using the log drain ID. For example:

```bash
terraform import aptible_log_drain.example-log-drain <ID>
```
