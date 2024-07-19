# Aptible Database Replica Resource

This resource is used to create and manage Database [Replicas and
Clusters](https://www.aptible.com/docs/core-concepts/managed-databases/managing-databases/replication-clustering)
running on Aptible Deploy.

!> Changing the handle of a replica will destroy the existing replica and
create a new one. It would then have to repopulate its contents from the
primary database

## Example Usage

```hcl
resource "aptible_replica" "example_database_replica" {
    env_id = 123
    primary_database_id = aptible_database.example_database.database_id
    handle = "example_database_replica"
    disk_size = 30
}
```

## Argument Reference

- `env_id` - The ID of the environment you would like to deploy your
  Database in. The Environment does not have to be the same as the primary
  database, but the Environment does have to be in the same
  [Stack](https://www.aptible.com/docs/core-concepts/architecture/stacks) as the
  primary Database. See main provider documentation for more on how to determine
  what you should use for `env_id`.
- `primary_database_id` - The ID of the Database the replica is being
  created from.
- `handle` - The handle for the Database. This must be all lower case, and
  only contain letters, numbers, `-`, `_`, or `.`
- `container_size` - (Default: 1024) The size of container used for the
  Database, in MB of RAM.
- `container_profile` - (Default: `m5`) Changes the CPU:RAM ratio of the
  Database container.
  - `m5` - General Purpose (1 CPU : 4 GB RAM)
  - `c5` - CPU Optimized (1 CPU : 2 GB RAM)
  - `r5` - Memory Optimized (1 CPU : 8 GB RAM)
- `disk_size` - The disk size of the Database, in GB.
- `iops` - The disk Input/Output Operations Per Second

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `replica_id` - The unique ID for the replica
- `default_connection_url` - The default
  [database credentials](https://www.aptible.com/docs/core-concepts/managed-databases/connecting-databases/database-credentials)
  in connection URL format

## Import

Existing Replica can be imported using the Replica ID. For example:

```bash
terraform import aptible_replica.example-replica <ID>
```
