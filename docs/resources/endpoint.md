# Aptible Endpoint Resource

This resource is used to create and manage Endpoints for
[Apps](https://www.aptible.com/docs/core-concepts/apps) and
[Databases](https://www.aptible.com/docs/core-concepts/managed-databases)
running on Aptible.

## Example Usage

### Simple Default Domain Endpoint

```hcl
resource "aptible_endpoint" "example" {
  env_id         = data.aptible_environment.example.env_id
  resource_id    = aptible_app.example.app_id
  # alternatively, one could also set "container_ports = [3000,3001]"
  container_port = 3000
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

## Endpoint Settings Example

Use the individual settings attributes to configure endpoint-level options for
an Application Endpoint. Supported settings vary by platform and endpoint type;
refer to the [Endpoint Documentation](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/overview)
for the type of Endpoint you are managing.

```hcl
resource "aptible_endpoint" "example_settings" {
  env_id         = data.aptible_environment.example.env_id
  resource_id    = aptible_app.example.app_id
  resource_type  = "app"
  process_type   = "cmd"
  endpoint_type  = "https"
  default_domain = true
  platform       = "alb"

  idle_timeout          = 120
  force_ssl             = true
  maintenance_page_url  = "https://example.com/maintenance"
}
```

## Argument Reference

- `env_id` - (Required) The ID of the environment you would like to deploy your
  Endpoint in. See main provider documentation for more on how to determine what
  you should use for `env_id`.
- `endpoint_type` - (Required) The type of Endpoint. Valid options are `https`,
  `tls`, or `tcp`. `tcp` should be used with `resource_type` of `database`.
- `resource_type` - (Required) The type of resource you are adding the Endpoint
  to. Valid options are `app` or `database`.
- `resource_id` - (Required) The ID of the resource you are adding the Endpoint
  to.
- `process_type` - (Required for Apps) The name of the service the Endpoint
  is for. See main provider documentation for more information on how to
  determine the service name.
- `container_port` - (Optional, App only) The port on the container which
  the Endpoint should forward traffic to. Mutually exclusive from
  `container_ports`. You should use this for `https` endpoints.
- `container_ports` - (Optional, App only) The ports in array form on the
  container which the Endpoint should forward traffic to. Mutually exclusive
  from `container_port`.
  Multiple container ports are only allowed on a `tcp` or `tls` endpoint.
- `default_domain` - (App only, Default: false) If the Endpoint should use the
  App's default `on-aptible.com` domain. Only one Endpoint per App can use the
  default domain. Cannot be used with `managed`.
- `managed` - (App only, Default: false) If Aptible should manage the HTTPS
  certificate for the Endpoint using the `custom_domain`. Cannot be used with
  `default_domain`.
- `domain` - (Optional, App only) Required when using Managed TLS (`managed`).
  The managed TLS Hostname the Endpoint should use.
- `internal` - (Default: false) If Endpoint should be available
  [internally or externally](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/overview#endpoint-placement)
  . Changing this will force the resource to be recreated.
- `platform` - (Default: `alb`) What type of
  [load balancer](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/https-endpoints/alb-elb)
  the Endpoint should use. Valid options are `alb` or `elb`. `resource_type` of
  `database` should use `elb`.
- `ip_filtering` - (Optional) The list of IPv4 CIDRs that the Endpoint will
  allow traffic from. If not provided, the Endpoint will not filter traffic. See
  the [IP Filtering](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/ip-filtering)
  documentation for more details.
- `shared` - (Optional, App only) If set, use shared load balancer resources
  with other apps on the same stack. Shared endpoints can only be used if your
  clients support SNI (most modern clients do) and you either use a default
  domain or an exact (non-wildcard) custom domain.
- `load_balancing_algorithm_type` - (Optional, ALB endpoints only) Determines which algorithm to use for
  [request routing](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/https-endpoints/overview#traffic). Valid options are `round_robin`, `least_outstanding_requests`, and `weighted_random`. The default is `round_robin`.

### Endpoint Settings

The following optional attributes configure endpoint-level behaviour. Omitting
an attribute leaves the platform default in place; removing a previously-set
attribute clears it back to the platform default on the next `apply`. Not all
settings are supported on every endpoint type — invalid combinations are caught
at plan time.

- `force_ssl` - (Optional, HTTPS endpoints only) When `true`, HTTP requests are
  redirected to HTTPS.
- `maintenance_page_url` - (Optional, HTTPS endpoints only) URL of a page to
  display when the endpoint returns a 503. Must be an `https://` URL.
- `idle_timeout` - (Optional) Connection idle timeout in seconds. Valid range:
  30–2400.
- `release_healthcheck_timeout` - (Optional, HTTPS endpoints only) Timeout in
  seconds for the release health check. Valid range: 1–900.
- `strict_health_checks` - (Optional, HTTPS endpoints only) When `true`, the
  load balancer uses strict health check settings.
- `show_elb_healthchecks` - (Optional, HTTPS endpoints only) When `true`,
  health check requests from the load balancer are visible in application logs.
- `ssl_protocols_override` - (Optional, HTTPS/TLS/gRPC endpoints only) Override
  the set of accepted TLS protocol versions. Valid values: `TLSv1 TLSv1.1 TLSv1.2`,
  `TLSv1 TLSv1.1 TLSv1.2 PFS`, `TLSv1.1 TLSv1.2`, `TLSv1.1 TLSv1.2 PFS`,
  `TLSv1.2`, `TLSv1.2 PFS`, `TLSv1.2 PFS TLSv1.3`, `TLSv1.3`. Values
  containing `PFS` are only valid on ALB endpoints.
- `ssl_ciphers_override` - (Optional, ELB/TLS/gRPC endpoints only) Override the
  cipher suite used for TLS negotiation.
- `disable_weak_cipher_suites` - (Optional, ELB/TLS/gRPC endpoints only) When
  `true`, weak cipher suites are disabled.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `endpoint_id` - The unique identifier for this Endpoint.
- `virtual_domain` - The public domain name that would correspond to the
  certificate served by this domain, if any.
- `external_hostname` - The public hostname of the load balancer serving this
  Endpoint.
- `dns_validation_record` - The CNAME record that needs to be created for
  Managed HTTPS to use
  [dns-01](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/managed-tls#dns-01)
  to verify ownership of the domain.
- `dns_validation_value` - The domain name to which the CNAME record should
  point for Managed HTTPS to use
  [dns-01](https://www.aptible.com/docs/core-concepts/apps/connecting-to-apps/app-endpoints/managed-tls#dns-01)
  to verify ownership of the domain.

## Import

Existing Endpoints can be imported using the Endpoint ID. For example:

```bash
terraform import aptible_endpoint.example-endpoint <ID>
```
