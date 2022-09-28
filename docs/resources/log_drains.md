# Aptible App Resource

This resource is used to create and manage
[Log drains](https://deploy-docs.aptible.com/docs/log-drains)
running on Aptible Deploy.

## Example Usage

Basic application deployment with configuration

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

Based on the `drain_type` it will change what arguments are required.

Because we support a lot of different destinations for log drains, we could not
properly validate all the required fields for each `drain_type` within
terraform.

It's entirely possible that terraform will pass its linting but the log drain isn't
necessarily set up properly in Aptible.

- `env_id` (Required) - The ID of the environment you would like to deploy your
  App in. See main provider documentation for more on how to determine what
  you should use for `env_id`.
- `handle` (Required) - The handle for the App. This must be all lower case, and
  only contain letters, numbers, `-`, `_`, or `.`
- `drain_type` (Required) - The type of log drain: `syslog_tls_tcp`,
  `elasticsearch_database`, `https_post`, `sumologic`, `logdna`, `datadog`,
  `papertrail`
- `drain_apps` - (Optional, default `True`) The configuration for sending logs from apps.
- `drain_databases` - (Optional, default `True`) The configuration for sending logs from databases.
- `drain_proxies` - (Optional, default `False`) The configuration for sending
  logs from proxies.
- `drain_ephemeral_sessions` - (Optional, default `False`) The configuration for sending logs from ephemeral sessions.
- `database_id` - The database id used for log drains
- `drain_host` - The host destination where log drains will be sent
- `drain_password` - The password for the host destination where logs
  will be sent
- `drain_port` - The port for the host destination where logs drains will be
  sent
- `logging_token` - Tag for syslog
- `url` - The destination url where the logs will be sent
- `tags` - Tags for logdna or datadog
- `token` - API token used for logdna or datadog
- `pipeline` - The elasticsearch pipeline to use

### Arguments based on `drain_type`

The following arguments are required regardless of the `drain_type`:

- `env_id`
- `handle`
- `drain_type`

#### `syslog_tls_tcp`

- `drain_host` (Required)
- `drain_port` (Required)

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
- `drain_password` (Optional)
- `tags` (Optional)

#### `papertrail`

- `drain_host` (Required)
- `drain_port` (Required)

## Attribute Reference

The following attributes are exported:

- `log_drain_id`
- `handle`
- `drain_type`
- `url`
- `logging_token`
- `drain_port`
- `drain_host`
- `drain_proxies`
- `drain_ephemeral_sessions`
- `drain_databases`
- `drain_apps`
- `env_id`
- `database_id`

## Import

Existing Apps can be imported using the Log Drain ID. For example:

```bash
terraform import aptible_log_drain.example-log-drain <ID>
```
