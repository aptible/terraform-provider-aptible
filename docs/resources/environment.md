# Aptible Environment Resource

This resource is used to create and manage [Environments](https://deploy-docs.aptible.com/docs/environments) running on Aptible Deploy.

## Example Usage

```hcl
resource "aptible_environment" "example_environment" {
  handle = "example"
  stack_id = 1
  org_id = "123e4567-e89b-12d3-a456-42661417400"  # insert your org uuid here!
}
```

## Argument Reference

- `stack_id` (Required) - The id of the [stack](https://deploy-docs.aptible.com/docs/stacks) you would like the environment to be provisioned on.  
- `org_id` (Required) - The id of the [organization](https://deploy-docs.aptible.com/docs/organizations) you would like the environment to be provisioned on
- `handle` - The handle for the environment.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `env_id` - The unique ID for the environment

## Import

Existing Environments can be imported using the Environment ID. For example:

```bash
terraform import aptible_environment.example <ID>
```
