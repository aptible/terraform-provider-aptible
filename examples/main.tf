// Examples of resources

#######################################################
# ENVIRONMENTS
#######################################################

data "aptible_environment" "example" {
    handle = "example"
}

#######################################################
# APPS
#######################################################

resource "aptible_app" "<app_handle>" {
  env_id = data.aptible_environment.example.env_id
  handle = "<app_handle>"
  config = {
      "APTIBLE_DOCKER_IMAGE" = "<docker_image>"
      "DATABASE_URL" = "<connection_url>"
      "ANOTHER_VAR" = "value"
  }
  service {
    process_type = "web"
    container_count = 3
    container_memory_limit = 2048
  }
  service {
    process_type = "background"
    container_count = 1
    container_memory_limit = 512
  }
}

#######################################################
# ENDPOINTS
#######################################################

resource "aptible_endpoint" "<endpoint_name>" {
  env_id = data.aptible_environment.example.env_id
  resource_id = resource.aptible_app.example.id
  resource_type = "app"             // other options: database
  type = "https"                    // other options: tcp, tls
  internal = true                   // or false for external
  container_port = 80               // port #
  ip_filtering = []                 // list of whitelisted IPs
  platform = "alb"                  // or "elb"
}

#######################################################
# DATABASES
#######################################################

resource "aptible_db" "<db_handle" {
  env_id = data.aptible_environment.example.env_id
  handle = "<db_handle>"
  db_type = "<db_type>"            // E.G. "postgresql", "mongodb", etc.
  container_size = "1024"
  disk_size = "10"
}

#######################################################
# REPLICAS
#######################################################

resource "aptible_replica" "<replica_handle" {
  env_id = data.aptible_environment.example.env_id
  primary_db_id = "<primary_db_id>"
  handle = "<replica_handle>"
}
