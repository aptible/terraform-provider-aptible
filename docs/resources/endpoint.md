# Aptible Endpoint Resource

This resource is used to create and manage Endpoints for
[Apps](https://www.aptible.com/documentation/deploy/reference/apps/endpoints.html)
and
[Databases](https://www.aptible.com/documentation/deploy/reference/databases/endpoints.html)
running on Aptible.

## Example Usage

### Simple Default Domain Endpoint

```hcl
resource "aptible_endpoint" "example" {
  env_id         = data.aptible_environment.example.env_id
  resource_id    = aptible_app.example.app_id
  resource_type  = "app"
  process_type   = "cmd"
  default_domain = true
}
```

## Managed Custom Domain

This example creates a Managed TLS Endpoint with a custom domain and uses AWS
Route53 to manage the Endpoint's DNS records:

```hcl
resource "aptible_endpoint" "example" {
  env_id        = data.aptible_environment.example.env_id
  resource_id   = aptible_app.example.app_id
  resource_type = "app"
  process_type  = "cmd"
  managed       = true
  domain        = "www.example.com"
}

data "aws_route53_zone" "example" {
  name = "example.com"
}

resource "aws_route53_record" "www" {
  zone_id = data.aws_route53_zone.example.zone_id
  name    = aptible_endpoint.example.domain
  type    = "CNAME"
  records = [aptible_endpoint.example.external_hostname]
}

resource "aws_route53_record" "dns01" {
  zone_id = data.aws_route53_zone.example.zone_id
  name    = aptible_endpoint.example.dns_validation_record
  type    = "CNAME"
  records = [aptible_endpoint.example.dns_validation_value]
}
```

## Argument Reference

- `env_id` - (Required) The ID of the environment you would like to deploy your
  Endpoint in. See main provider documentation for more on how to determine what
  you should use for `env_id`.
- `endpoint_type` - (Required) The type of Endpoint. Valid options are `https`,
  `tls`, or `tcp`.
- `resource_type` - (Required) The type of resource you are adding the Endpoint
  to. Valid options are `app` or `database`.
- `resource_id` - (Required) The ID of the resource you are adding the Endpoint
  to.
- `process_type` - (Required for Apps) The name of the service the Endpoint
  is for. See main provider documentation for more information on how to
  determine the sevice name.
- `container_port` - (Optional, App only) The port on the container which
  the Endpoint should forward traffic to.
- `default_domain` - (App only, Default: false) If the Endpoint should use the
  App's default `on-aptible.com` domain. Only one Endpoint per App can use the
  default domain. Cannot be used with `managed`.
- `managed` - (App only, Default: false) If Aptible should manage the HTTPS
  certificate for the Endpoint using the `custom_domain`. Cannot be used with
  `default_domain`.
- `domain` - (Optional, App only) Required when using Managed TLS (`managed`).
  The managed TLS Hostname the Endpoint should use.
- `internal` - (Default: false) If Endpoint should be available
  [internally or externally](https://deploy-docs.aptible.com/docs/endpoints#endpoint-placement)
  . Changing this will force the resource to be recreated.
- `platform` - (Default: `alb`) What type of 
  [load balancer](https://www.aptible.com/documentation/deploy/reference/apps/endpoints/https-endpoints/alb-elb.html#alb-elb)
  the Endpoint should use. Valid options are `alb` or `elb`.
- `ip_filtering` - (Optional) The list of IPv4 CIDRs that the Endpoint will
  allow traffic from. If not provided, the Endpoint will not filter traffic.
  See the [IP Filtering](https://deploy-docs.aptible.com/docs/ip-filtering)
  documentation for more details.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `endpoint_id` - The unique identifier for this Endpoint.
- `virtual_domain` - The public domain name that would correspond to the
  certificate served by this domain, if any.
- `external_hostname` - The public hostname of the load balancer serving this
  Endpoint.
- `dns_validation_record` - The CNAME record that needs to be created for
  Managed HTTPS to use
  [dns-01](https://deploy-docs.aptible.com/docs/managed-tls#dns-01) to verify
  ownership of the domain.
- `dns_validation_value` - The domain name to which the CNAME record should
  point for Managed HTTPS to use
  [dns-01](https://deploy-docs.aptible.com/docs/managed-tls#dns-01) to verify
  ownership of the domain.

## Import

Existing Endpoints can be imported using the Endpoint ID. For example:

```bash
terraform import aptible_endpoint.example-endpoint <ID>
```
