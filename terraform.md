# Terraform Provider for Aptible Deploy

To use the provider:

1. Run `go build -o terraform-provider-aptible`.
2. Add a config file in the root directory named `main.tf`.
3. Run `terraform init`.
4. Run `terraform plan`.
5. Run `terraform apply`. **Warning: This step will cause real resources to be created!**

## Resources

This provider has 4 resources: apps, databases, endpoints, and replicas.

### Apps

- Deployment
  - Currently, only direct docker image deployment is supported.
  - For more information on direct docker image deployment, refer to the
    [Aptible Deploy documentation](https://www.aptible.com/docs/core-concepts/apps/deploying-apps/image/deploying-with-docker-image/overview).
- Configuration
  - Configuration is done via setting environment variables.
  - For more information on configuration variables in Aptible Deploy, refer to
    the
    [Aptible Deploy documentation](https://www.aptible.com/docs/core-concepts/apps/deploying-apps/configuration).
  - Connecting to a database can be done using the variable `DATABASE_URL` and
    setting it to the database's connection URL.

### Databases

### Endpoints

### Replicas

---

## Data Sources

This provider has 1 data source: environments.

### Environments

- The environments data source gets the environment ID corresponding to a given
  handle.

---

## Configuration

You can find an example configuration file in `examples/main.tf`.

### Apps

Below are the attributes defined in the app config with the accepted values.

- **env_id (environment ID)**: use
  `data.aptible_environment.<environment_name>.env_id`. For example, if your
  environment is called `test`, then the config should contain:

  > `env_id = data.aptible_environment.test.env_id`

- **handle (app handle)**: the handle for your app.
   For example, if your app is called `test`, then the config should contain:

  > `handle = "test"`

- **config**: environment variables for your app (e.g., `DATABASE_URL`).
- **settings**: App settings. To deploy a Docker image, set `APTIBLE_DOCKER_IMAGE` here.
- **sensitive_settings**: App settings treated as sensitive by Terraform (e.g., private registry credentials).

For example, if you wanted to deploy using the `nginx` image, then the config should contain:

> `settings = {"APTIBLE_DOCKER_IMAGE" = "nginx"}`

The `app_id` and `git_repo` attributes will be generated upon running
`terraform apply`.

### Databases

Below are the attributes defined in the database config with the accepted values.

- **env_id**: the ID for your environment.
- **handle**: the handle for your database.
- **db_type**: the type of your database.
- **container_size**: the container size for your database.
- **disk_size**: the disk size for your database.

The `db_id` and `default_connection_url` attributes will be generated upon
running `terraform apply`.

### Endpoints

Below are the attributes defined in the endpoint config with the accepted values.

### Replicas

Below are the attributes defined in the replica config with the accepted values.

### Environments

Below are the attributes defined in the environment config with the accepted
values.

- **handle**: the handle for your environment.
  For example, if your environment is called `test`, then the config should
  contain:
  > `handle = "test"`

---

## Acceptance Tests

This provider has acceptance tests to test basic functionality for each
resource, as well as tests to ensure that invalid inputs return errors.

To run the acceptance tests:

1. Set the `APTIBLE_ENVIRONMENT_ID` to the test environment ID to run acceptance
   tests against. This will create real AWS resources.
2. Make sure your `APTIBLE_AUTH_ROOT_URL`, `APTIBLE_API_ROOT_URL`,
   and `APTIBLE_ACCESS_TOKEN` environment variables are set.
3. Run `make testacc`.

With the default parallelization, the test take about 15 minutes to run.
