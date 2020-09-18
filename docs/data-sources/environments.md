# Environment Data Source

[Environments](https://www.aptible.com/documentation/deploy/reference/environments.html)
are where all resources are created and managed. Currently the Aptible Deploy Terraform
provider does not manage Environments, but this data source allows you to look-up the
IDs of existing Environments for use within Terraform.

## Example Usage

### Determining the Environment ID

If you have an Environment with the handle "techco-test-environment" you can
create the data source:

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

## Argument Reference

- `handle` - The handle for the Environment. This must be all lower case, and
  only contain letters, numbers, `-`, `_`, or `.`

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique ID for an Environment suitable for use in `env_id` attributes
