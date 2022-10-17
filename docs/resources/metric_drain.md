# Aptible Metric Drain Resource

This resource is used to create and manage
[Metric Drains](https://deploy-docs.aptible.com/docs/metric-drains)
running on Aptible Deploy.

## Example Usage

```hcl
resource "aptible_metric_drain" "influxdb_database_metric_drain" {
  env_id      = data.aptible_environment.example.env_id
  database_id = aptible_database.example.database_id
  drain_type  = "influxdb_database"
  handle      = "aptible-hosted-metric-drain"
}
```

```hcl
resource "aptible_metric_drain" "influxdb_metric_drain" {
  env_id     = data.aptible_environment.example.env_id
  drain_type = "influxdb"
  handle     = "influxdb-metric-drain"
  url        = "https://influx.example.com:443"
  username   = "example_user"
  password   = "example_password"
  database   = "metrics"
}
```

```hcl
resource "aptible_metric_drain" "datadog_metric_drain" {
  env_id     = data.aptible_environment.example.env_id
  drain_type = "datadog"
  api_key    = "xxxxx-xxxxx-xxxxx"
}
```

## Argument Reference

- `env_id` (Required) - The ID of the environment you would like to create the
  metric drain in. See main provider documentation for more on how to determine
  what you should use for `env_id`.
- `handle` (Required) - The handle for the metric drain. This must be all lower
  case, and only contain letters, numbers, `-`, `_`, or `.`.
- `drain_type` (Required) - The type of metric drain: `influxdb_database`,
  `influxdb`, `datadog`
- `database_id` - The ID of the Aptible InfluxDB database for
  `influxdb_database` drains to send metrics to.
- `url` - The URL (scheme, host, and port) for `influxdb` drains to send metrics
  to.
- `username` - The user for `influxdb` drains to use for authentication.
- `password` - The password for `influxdb` drains to use for authentication.
- `database` - The 
  [InfluxDB database](https://docs.influxdata.com/influxdb/v1.8/concepts/glossary/#database)
  for `influxdb` drains to send the metrics to.
- `api_key` - The API key for `datadog` drains to use for authentication.
- `series_url` - The series API URL for `datadog` drains to send metrics to.
  Examples: `https://app.datadoghq.com/api/v1/series`,
  `https://us3.datadoghq.com/api/v1/series`,
  `https://app.datadoghq.eu/api/v1/series`,
  `https://app.ddog-gov.com/api/v1/series`

### Arguments based on `drain_type`

All `aptible_metric_drain` resources require the following attributes:

- `env_id`
- `handle`
- `drain_type`

#### `influxdb_database`

- `database_id` (Required)

#### `influxdb`

- `url` (Required)
- `username` (Required)
- `password` (Required)
- `database` (Required)

#### `datadog`

- `api_key` (Required)
- `series_url` (Optional)

## Attribute Reference

In addition to all the arguments listed above, the following attributes are 
exported:

- `metric_drain_id` - The unique ID for the metric drain.

## Import

Existing metric drains can be imported using the metric drain ID. For example:

```bash
terraform import aptible_metric_drain.example <ID>
```
