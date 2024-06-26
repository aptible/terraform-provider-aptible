# Backup Retention Policy Data Source

An Environment's
[Backup Retention Policy](https://www.aptible.com/docs/core-concepts/managed-databases/managing-databases/database-backups#managing-backup-retention-policy)
controls how automated backups are retained and destroyed.

## Example Usage

```hcl
data "aptible_environment" "example" {
	handle = "example-env"
}

data "aptible_backup_retention_policy" "example" {
    env_id = data.aptible_environment.example.env_id
}
```

## Argument Reference

- `env_id` (Required) - The ID of the environment to retrieve the backup
  retention policy for.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `daily` - The number of daily backups to retain per database.
- `monthly` - The number of monthly backups to retain per database.
- `yearly` - The number of yearly backups to retain per database.
- `make_copy` - Whether backups should be copied to another region.
- `keep_final` - Whether the final backup of databases should be retained when
  they're deleted.
