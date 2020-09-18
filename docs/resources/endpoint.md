# Aptible App Resource

This resource is used to create and manage endpoints for
[Apps](https://www.aptible.com/documentation/deploy/reference/apps/endpoints.html)
and
[Databases](https://www.aptible.com/documentation/deploy/reference/databases/endpoints.html)
running on Aptible.

## Example Usage

```hcl
resource "aptible_endpoint" "example_endpoint" {
    env_id = 123
    process_type = "cmd"
    resource_id = aptible_app.example_app.app_id
    default_domain = true
    endpoint_type = "https"
    internal = false
    platform = "alb"
    container_port = 5000
    managed = true
}
```

## Argument Reference

- `env_id` - The ID of the environment you would like to deploy your
  App in. See main provider documentation for more on how to determine what
  you should use for `env_id`.
- `container_port` - (Optional, Apps only) The port on the container which
  the Endpoint should forward traffic to.
- `default_domain` - (Optional, App only) Whether or not we should create
  a domain using our default on-aptible.com URL for the Endpoint. Only one
  Endpoint using a default domain is permitted per App.
- `endpoint_type` - The type of Endpoint. Valid options are `https` or
  `tcp`.
- `internal` - (Optional) Whether the Endpoint should be available
  exclusively internally. Default is `false`.
- `managed` - (Optional, App only) Whether or not Aptible should manage
  the HTTPS certificate for the Endpoint.
- `platform` - (Optional) Whether to use an [ALB or ELB](https://www.aptible.com/documentation/deploy/reference/apps/endpoints/https-endpoints/alb-elb.html#alb-elb).
  Supported values are `alb` or `elb`. Default is `elb`.
- `resource_id` - The ID of the resource you are adding an endpoint to.
- `resource_type` - The type of resource you are adding an endpoint to.
  This should be either `app` or `database`.
- `process_type` - (Required for Apps) The name of the service the Endpoint
  is for. See main provider documentation for more information on how to
  determine the sevice name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `endpoint_id` - The unique identifier for this Endpoint
- `virtual_domain` - The public domain name that would correspond to the
  certificate served by this domain, if any

## Import

Existing Endpoints can be imported using the Endpoint ID. For example:

```bash
terraform import aptible_endpoint.example-endpoint <ID>
```
