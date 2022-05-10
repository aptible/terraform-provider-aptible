# This deploys our Aptible Demo App
# https://www.aptible.com/documentation/deploy/tutorials/deploy-demo-app.html

terraform {
  required_providers {
    aptible = {
      source  = "aptible/aptible"
      version = "~>0.1"
#      The token may be set here, but proceed with extreme caution. This is not encouraged. You should use the
#      environment variable `APTIBLE_ACCESS_TOKEN` or the CLI `aptible login`
#      token = <SOME TOKEN HERE>
    }
  }
}

# TODO: Enter your account handle here
data "aptible_environment" "demo" {
  handle = ""
}

resource "aptible_app" "demo-app" {
  env_id = data.aptible_environment.demo.env_id
  handle = "demo-app"
  config = {
    "APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/deploy-demo-app"
    "DATABASE_URL"         = aptible_database.demo-pg.default_connection_url
    "REDIS_URL"            = aptible_database.demo-redis.default_connection_url
  }
  service {
    process_type           = "web"
    container_count        = 2
    container_memory_limit = 512
  }
  service {
    process_type           = "background"
    container_count        = 0
    container_memory_limit = 512
  }

  # https://www.aptible.com/documentation/deploy/tutorials/deploy-demo-app.html#run-database-migrations
  # This is a one-time execution at application creation for demo purposes.
  # If you need to run migrations on each app release, you should use a before_release command
  # https://www.aptible.com/documentation/deploy/reference/apps/aptible-yml.html#before-release
  provisioner "local-exec" {
    command = "aptible ssh --app demo-app python migrations.py"
  }
}

resource "aptible_database" "demo-pg" {
  env_id         = data.aptible_environment.demo.env_id
  handle         = "demo-pg"
  database_type  = "postgresql"
  container_size = 512
  disk_size      = 10
}

resource "aptible_database" "demo-redis" {
  env_id         = data.aptible_environment.demo.env_id
  handle         = "demo-redis"
  database_type  = "redis"
  container_size = 512
  disk_size      = 10
}

resource "aptible_endpoint" "demo-app-public-endpoint" {
  env_id         = data.aptible_environment.demo.env_id
  resource_type  = "app"
  process_type   = "web"
  resource_id    = aptible_app.demo-app.app_id
  default_domain = true
  endpoint_type  = "https"
  container_port = 5000
}
