# Aptible Provider

## Feature Support

For a detailed list of what features the provider supports see the
[Feature Support Matrix](https://www.aptible.com/docs/reference/interface-feature)
in the Aptible Documentation.

## Example Usage

### Authentication and Authorization

Authorization and Authentication is controlled using the same mechanism
that the [CLI](https://www.aptible.com/docs/reference/aptible-cli/overview) uses.
Therefore, you should log into the account you want to use Terraform with using
the `aptible login` CLI command before running any Terraform commands.

As another option the environment variables `APTIBLE_USERNAME` and
`APTIBLE_PASSWORD` can be set for the provider to use. In this case it is
strongly recommended that a robot account be used, especially as MFA needs to
be disabled for truly automated runs.

### Determining the Environment ID

Each resource managed via Terraform requires an Environment ID specifying which
[Environment](https://www.aptible.com/docs/core-concepts/architecture/environments)
the resource should be created in. Currently the Aptible Deploy Terraform
provider does not manage Environments, so you will need the Environment ID for
a pre-existing Environment. The easiest way to determine the Environment ID is
by using the Environment data source. For example, if you have an Environment
with the handle "techco-test-environment" you can create the data source:

```hcl
data "aptible_environment" "techco-test-environment" {
    handle = "techco-test-environment"
}
```

Once defined, you can use this data source in your resource definitions.
For example, when defining an App:

```hcl
data "aptible_app" "techo-app" {
    env_id = data.aptible_environment.techco-test-environment.env_id
    handle = "techo-app"
}
```

You can also specify the environment as a resource. You will need to pass in an
`org_id` (ex: `dashboard.aptible.com/organizations/<ORG_ID>/members`) and
`stack_id` (ex: `dashboard.aptible.com/stack/<STACK_ID>/accounts`) which you can
get from the Aptible dashboard (on the settings/members panel (`org_id`) or on
the stack view pages (`stack_id`)):

```hcl
resource "aptible_environment" "techco-test-environment" {
    org_id = "some-uuid-that-represents-your-org"
    stack_id = "your_stack_id"
    handle = "techco-test-environment"
}
```

### Apps

[Apps](https://www.aptible.com/docs/core-concepts/apps) can be
created using the `terraform_aptible_app` resource.

```hcl
data "aptible_app" "APP" {
    handle = "APP_HANDLE"
}
```

#### Configuring and Deploying Apps

!> Currently the only supported deployment method via Terraform is of
Docker images hosted in a Docker image registry.

Application environment variables can be managed via the nested `config` element.

```hcl
resource "aptible_app" "APP" {
    env_id = ENVIRONMENT_ID
    handle = "APP_HANDLE"
    config = {
        "KEY" = "value"
    }
}
```

To deploy a Docker image, set `APTIBLE_DOCKER_IMAGE` in `settings`.
Authentication for Docker images located in private repositories can be provided
using `APTIBLE_PRIVATE_REGISTRY_USERNAME` and
`APTIBLE_PRIVATE_REGISTRY_PASSWORD` in `sensitive_settings`.

```hcl
resource "aptible_app" "APP" {
    env_id = ENVIRONMENT_ID
    handle = "APP_HANDLE"
    settings = {
        "APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/deploy-demo-app"
    }
    sensitive_settings = {
        "APTIBLE_PRIVATE_REGISTRY_USERNAME" = "registry_username"
        "APTIBLE_PRIVATE_REGISTRY_PASSWORD" = "registry_password"
    }
}
```

#### Scaling Services

Each App is comprised of one or more
[Services](https://www.aptible.com/docs/core-concepts/apps/deploying-apps/services).
These Services must be defined in the
[Procfile](https://www.aptible.com/docs/how-to-guides/app-guides/define-services#explicit-services-procfiles)
for your App.

Services can be scaled independently both in terms of the number of running
[containers](https://www.aptible.com/docs/core-concepts/architecture/containers/overview)
and size of the running Containers. This is done using the nested `service`
element for the App resource:

-> The `process_type` in the `service` element maps directly to the
Service name used in the Procfile. If you are not using a Procfile,
you will have a single Service with the `process_type` of `cmd`

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
    }
}
```

#### Autoscaling

Services can be configured to [Autoscale](https://www.aptible.com/docs/core-concepts/scaling/app-scaling)
either vertically or horizontally. The configuration is per service

```hcl
resource "aptible_app" "APP" {
    env_id = ENVIRONMENT_ID
    handle = "APP_HANDLE"
    service {
        process_type = "SERVICE_NAME1"
        container_count = 1
        container_memory_limit = 1024
        autoscaling_policy {
            autoscaling_type = "horizontal"
            min_containers = 2
            max_containers = 5
            min_cpu_threshold = 0.4
            max_cpu_threshold = 0.8
        }
    }
    service {
        process_type = "SERVICE_NAME2"
        container_count = 2
        container_memory_limit = 2048
        autoscaling_policy {
            autoscaling_type = "vertical"
            mem_scale_up_threshold = 0.8
        }
    }
}
```

### Endpoints

Endpoints for [Apps](https://www.aptible.com/docs/core-concepts/apps) and
[Databases](https://www.aptible.com/docs/core-concepts/managed-databases) can be
managed using the `terraform_aptible_endpoint` resource.

```hcl
resource "aptible_endpoint" "EXAMPLE" {
    env_id = ENVIRONMENT_ID
    process_type = "SERVICE_NAME"
    resource_id = aptible_app.APP.app_id
    default_domain = true
    endpoint_type = "https"
    internal = false
    platform = "alb"
    container_port = 5000
}
```

### Databases

[Databases](https://www.aptible.com/docs/core-concepts/managed-databases) can be
managed using the `terraform_aptible_database` resource.

```hcl
resource "aptible_database" "DATABASE" {
    env_id = ENVIRONMENT_ID
    handle = "DATABASE_HANDLE"
    database_type = "redis"
    container_size = 512
    disk_size = 10
}
```

#### Replication

Database [Replicas and
Clusters](https://www.aptible.com/docs/core-concepts/managed-databases/managing-databases/replication-clustering)
can be created using the `terraform_aptible_replica` resource.

```hcl
resource "aptible_replica" "REPLICA_HANDLE" {
    env_id = ENVIRONMENT_ID
    primary_database_id = aptible_database.DATABASE.database_id
    handle = "REPLICA_HANDLE"
    disk_size = 30
}
```

## Argument Reference

There are currently no arguments to provide directly to the provider
