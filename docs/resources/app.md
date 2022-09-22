# Aptible App Resource

This resource is used to create and manage
[Apps](https://www.aptible.com/documentation/deploy/reference/apps.html)
running on Aptible Deploy.

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

- `process_type` - The `process_type` maps directly to the Service name used in
  the Procfile. If you are not using a Procfile, you will have a single Service
  with the `process_type` of `cmd`
- `container_count` - (Optional) The number of unique containers running the
  service for horizontal scaling
- `container_memory_limit` - (Optional) Increase the memory limit of each
  container in the service for vertical scaling

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `app_id` - The unique ID of the application
- `git_repo` - The git remote associated with the application

## Import

Existing Apps can be imported using the App ID. For example:

```bash
terraform import aptible_app.example-app <ID>
```
