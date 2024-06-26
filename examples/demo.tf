# This deploys our Aptible Demo App
# https://www.aptible.com/docs/getting-started/deploy-starter-template/python-flask

terraform {
  required_providers {
    aptible = {
      source  = "aptible/aptible"
      version = "~>0.1"
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
    container_count        = 1
    container_memory_limit = 512
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
