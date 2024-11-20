# Aptible App Resource

This resource is used to create and manage
[Apps](https://www.aptible.com/docs/core-concepts/apps) running on Aptible
Deploy.

## Example Usage

Basic application deployment with configuration

```hcl
resource "aptible_app" "example_app" {
    env_id = 123
    handle = "example_app"
    config = {
        "KEY" = "value"
        "APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/deploy-demo-app"
        "APTIBLE_PRIVATE_REGISTRY_USERNAME" = "registry_username"
        "APTIBLE_PRIVATE_REGISTRY_PASSWORD" = "registry_password"
    }
}
```

Application with defined services to control scaling through Terraform

```hcl
resource "aptible_app" "APP" {
    env_id = ENVIRONMENT_ID
    handle = "APP_HANDLE"
    service {
        process_type = "SERVICE_NAME1"
        container_count = 1
        container_memory_limit = 1024
    }
    service {
        process_type = "SERVICE_NAME2"
        container_count = 2
        container_memory_limit = 2048
        force_zero_downtime = true
        service_sizing_policy {
          autoscaling_type = "horizontal"
          min_containers = 2
          max_container = 5
          min_cpu_threshold = 0.4
          max_cpu_threshold = 0.8
        }
    }
}
```

## Argument Reference

- `env_id` - The ID of the environment you would like to deploy your
  App in. See main provider documentation for more on how to determine what
  you should use for `env_id`.
- `handle` - The handle for the App. This must be all lower case, and
  only contain letters, numbers, `-`, `_`, or `.`
- `config` - (Optional) The configuration for the App. This should be a
  map of `KEY = VALUE`.
- `service` - (Optional) A block to manage scaling for services. See the main
  provider docs for additional details.

The `service` block supports:

- `process_type` - (Default: `cmd`) The `process_type` maps directly to the
  Service name used in the Procfile. If you are not using a Procfile, you will
  have a single Service with the `process_type` of `cmd`.
- `container_count` - (Default: 1) The number of unique containers running the
  service.
- `container_memory_limit` - (Default: 1024) The memory limit (in MB) of the
  service's containers.
- `container_profile` - (Default: `m5`) Changes the CPU:RAM ratio of the
  service's containers.
  - `m4` - General Purpose (1 CPU : 4 GB RAM)
  - `c5` - CPU Optimized (1 CPU : 2 GB RAM)
  - `r5` - Memory Optimized (1 CPU : 8 GB RAM)
- `force_zero_downtime` - (Default: false) For services without endpoints, force
  a zero-downtime release and leverage docker healthchecks for the containers. Please
  note that docker healthchecks are required unless `simple_health_check` is enabled.
  [For more information please see the docs](https://www.aptible.com/docs/core-concepts/apps/deploying-apps/releases/overview).
- `simple_health_check` - (Default: false) For services without endpoints, if
  force_zero_downtime is enabled, do a simple uptime check instead of using docker healthchecks.
- `service_sizing_policy` - (Optional) A block to manage autoscaling for services. See
  the main provider docs for additional details.

The `service_sizing_policy` block supports:

- `autoscaling_type` - The type of autoscaling. Must be either `horizontal` or `vertical`.
- `metric_lookback_seconds` - (Default: 1800) The duration in seconds for 
  retrieving past performance metrics.
- `percentile` - (Default: 99) The percentile for evaluating metrics.
- `post_scale_up_cooldown_seconds` - (Default: 60) The waiting period in seconds after an automated
 scale-up before another scaling action can be considered.
- `post_scale_down_cooldown_seconds` - (Default: 300) The waiting period in seconds after an automated
 scale-down before another scaling action can be considered.
- `post_release_cooldown_seconds` - (Default: 300) The time in seconds to wait following a
  deploy before another scaling action can be considered.
- `mem_cpu_ratio_r_threshold` - (Default: 4.0) Establishes the ratio of Memory (in GB) to CPU (in CPUs)
  at which values exceeding the threshold prompt a shift to an R (Memory Optimized) profile.
- `mem_cpu_ratio_c_threshold` - (Default: 2.0) Sets the Memory-to-CPU ratio threshold,
  below which the service is transitioned to a C (Compute Optimized) profile.
- `mem_scale_up_threshold` - (Default: 0.9) Vertical autoscaling only - Specifies the percentage
  of the current memory limit at which the service’s memory usage triggers an up-scaling action.
- `mem_scale_down_threshold` - (Default: 0.75) Vertical autoscaling only - Specifies the percentage
  of the current memory limit at which the service’s memory usage triggers a down-scaling action.
- `minimum_memory` - (Default: 2048) Vertical autoscaling only - Sets the lowest memory
  limit to which the service can be scaled down by Autoscaler.
- `maximum_memory` - Vertical autoscaling only - Defines the upper memory threshold,
  capping the maximum memory allocation possible through Autoscaler. If blank,
  the container can scale to the largest size available.
- `min_cpu_threshold` - Horizontal autoscaling only - Specifies the percentage of the current CPU usage at which
  a down-scaling action is triggered.
- `max_cpu_threshold` - Horizontal autoscaling only - Specifies the percentage of the current CPU usage at which
  an up-scaling action is triggered.
- `min_containers` - Horizontal autoscaling only - Sets the lowest container count to which
  the service can be scaled down by Autoscaler.
- `max_containers` - Horizontal autoscaling only - Sets the highest container count to which
  the service can be scaled up to by Autoscaler.
- `scale_up_step` - (Default: 1) Horizontal autoscaling only - Sets the amount of containers to add
  when autoscaling (ex: a value of 2 will go from 1->3->5). Container count will never exceed the configured maximum.
- `scale_down_step` - (Default: 1) Horizontal autoscaling only - Sets the amount of containers to remove when
  autoscaling (ex: a value of 2 will go from 4->2->1). Container count will never exceed the configured minimum.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `app_id` - The unique ID of the application.
- `git_repo` - The git remote associated with the application.

## Import

Existing Apps can be imported using the App ID. For example:

```bash
terraform import aptible_app.example-app <ID>
```
