# Aptible Database Resource

This resource is used to create and manage
[Databases](https://www.aptible.com/docs/core-concepts/managed-databases)
running on Aptible Deploy.

!> Changing the handle of a database will destroy the existing database and
create a new one, resulting in a database without data. The old database can
still be recovered by [restoring a
backup](https://www.aptible.com/docs/core-concepts/managed-databases/managing-databases/database-backups#restoring-from-a-backup)
as long as your retention policy supports final backups.

## Example Usage

```hcl
resource "aptible_database" "example_database" {
    env_id = 123
    handle = "example_database"
    database_type = "redis"
    version = ""
    container_size = 512
    disk_size = 10
}
```

## Argument Reference

- `env_id` - The ID of the environment you would like to deploy your
  Database in. See main provider documentation for more on how to determine what
  you should use for `env_id`.
- `handle` - The handle for the Database. This must be all lower case, and
  only contain letters, numbers, `-`, `_`, or `.`
- `database_type` - The type of Database.
- `version` - (Optional) The version of the Database. If none is specified,
  this defaults to the latest recommended version.
- `container_size` - (Default: 1024) The size of container used for the
  Database, in MB of RAM.
- `container_profile` - (Default: `m5`) Changes the CPU:RAM ratio of the
  Database container.
  - `m5` - General Purpose (1 CPU : 4 GB RAM)
  - `c5` - CPU Optimized (1 CPU : 2 GB RAM)
  - `r5` - Memory Optimized (1 CPU : 8 GB RAM)
- `disk_size` - The disk size of the Database, in GB.
- `iops` - The disk Input/Output Operations Per Second
- `enable_backups` - (Default: `true`) Whether to automatically backup the database according to the retention policy.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `database_id` - The unique ID for the database
- `database_image_id` - The image used for running the database. Normally only
  used for support or debugging purposes
- `default_connection_url` - The default
  [database credentials](https://www.aptible.com/docs/core-concepts/managed-databases/connecting-databases/database-credentials)
  in connection URL format
- `connection_urls` - A list of all available database credentials in connection
  URL format

## Import

Existing Databases can be imported using the Database ID. For example:

```bash
terraform import aptible_database.example-database <ID>
```
