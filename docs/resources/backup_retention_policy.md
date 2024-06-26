# Backup Retention Policy Resource

This resource is used to manage an Aptible Environment's
[Backup Retention Policy](https://www.aptible.com/docs/core-concepts/managed-databases/managing-databases/database-backups#managing-backup-retention-policy).

## Example Usage

```hcl
data "aptible_environment" "example" {
	handle = "example-env"
}

resource "aptible_backup_retention_policy" "example" {
  env_id     = data.aptible_environment.example.env_id
  daily      = 30
  monthly    = 12
  yearly     = 6
  make_copy  = false
  keep_final = true
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
  name     = "short-lived-env"
}

resource "aptible_backup_retention_policy" "example" {
  env_id     = aptible_environment.example.env_id
  daily      = 7
  monthly    = 0
  yearly     = 0
  make_copy  = false
  keep_final = false
}
```

## Argument Reference

!> Reducing the number of backups retained by an environment's backup retention
policy, including turning off boolean options, will cause existing, automated
backups that no longer fit the criteria to be deleted the next time the policy
is evaluated. Deleted backups cannot be recovered and the new policy may violate
your organization's internal compliance controls so always make sure to review
the `terraform plan` before applying.

~> Each environment has a single backup retention policy so there should only be
one policy resource per environment. Defining multiple policy resources for an
environment could lead to unexpected behavior.

- `env_id` (Required) - The ID of the environment you would like to manage the
  backup retention policy of. See main provider documentation for more on how to determine
  what you should use for `env_id`.
- `daily` (Required) - The number of daily backups to retain per database.
  Minimum of 1.
- `monthly` (Required) - The number of monthly backups to retain per database.
- `yearly` (Required) - The number of yearly backups to retain per database.
- `make_copy` (Required) - Whether backups should be copied to another region.
- `keep_final` (Required) - Whether the final backup of databases should be
  retained when they're deleted.

## Attribute Reference

In addition to all the arguments listed above, the following attributes are
exported:

- `policy_id` - The unique ID for the backup retention policy.

## Import

Existing backup retention policies can be imported using the environment ID.
For example:

```bash
terraform import aptible_backup_retention_policy.example <ENV_ID>
```
