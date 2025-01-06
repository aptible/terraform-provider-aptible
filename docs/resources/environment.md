# Aptible Environment Resource

This resource is used to create and manage [Environments](https://www.aptible.com/docs/core-concepts/architecture/environments) running on Aptible Deploy.

## Example Usage

```hcl
data "aptible_stack" "example" {
	name = "example-stack"
}

resource "aptible_environment" "example" {
  stack_id = data.aptible_stack.example.stack_id
  org_id   = data.aptible_stack.example.org_id
  handle     = "example-env"
}
```

### Short-lived Environment

An environment that has backups associated with it cannot be deleted. Most automated backups are deleted along with their database but final backups will be retained if the environment is configured for it. Setting `keep_final = false` allows short-lived environments to be easily torn down with `terraform destroy`. Reducing the number of retained backups can also reduce the cost of databases in these environments.

```hcl
data "aptible_stack" "example" {
	name = "example-stack"
}

resource "aptible_environment" "example" {
  stack_id = data.aptible_stack.example.stack_id
  org_id   = data.aptible_stack.example.org_id
  handle     = "short-lived-env"

  backup_retention_policy {
    daily      = 1
    monthly    = 0
    yearly     = 0
    make_copy  = false
    keep_final = false
  }
}
```

## Argument Reference

- `stack_id` (Required) - The id of the [stack](https://www.aptible.com/docs/core-concepts/architecture/stacks) you would like the environment to be provisioned on.
- `org_id` (Optional) - The id of the [organization](https://www.aptible.com/docs/core-concepts/security-compliance/access-permissions#organization) you would like the environment to be provisioned on. If the `org_id` is not provided, the provider will attempt to determine it for you. If you are only a member of a single Aptible organization or the environment is on a dedicated stack, it will certainly be able to.
- `handle` (Required) - The handle for the environment.
- `backup_retention_policy` - (Optional) A block defining the environment's backup retention policy. An environment may only have one policy block.

The `backup_retention_policy` block supports:

!> Reducing the number of backups retained by an environment's backup retention policy, including turning off boolean options, will cause existing, automated backups that do not fit the new criteria to be deleted. Deleted backups cannot be recovered and the new policy may violate your organization's internal compliance controls. Always make sure to review the `terraform plan` before applying to ensure the policy is properly defined.

- `daily` (Required) - The number of daily backups to retain per database. Minimum of 1.
- `monthly` (Required) - The number of monthly backups to retain per database.
- `yearly` (Required) - The number of yearly backups to retain per database.
- `make_copy` (Required) - Whether backups should be copied to another region.
- `keep_final` (Required) - Whether the final backup of databases should be retained when they're deleted.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `env_id` - The unique ID for the environment

## Import

Existing Environments can be imported using the Environment ID. For example:

```bash
terraform import aptible_environment.example <ID>
```
