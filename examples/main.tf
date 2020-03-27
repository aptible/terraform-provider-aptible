// Examples of resources

#######################################################
# APPS
#######################################################

resource "aptible_app" "<app_handle>" {
    env_id = "<env_id>"
    handle = "<app_handle>"
    config = {
        "APTIBLE_DOCKER_IMAGE" = "<docker_image>"
        "DATABASE_URL" = "<connection_url>"
        "ANOTHER_VAR" = "value"
    }
}

#######################################################
# ENDPOINTS
#######################################################

resource "aptible_endpoint" "<endpoint_name>" {
  env_id = "<env_id>"
  app_id     = "<app_id>"
  type = "HTTPS"                    // other options: TCP, TLS
  internal = true                   // or false for external
  container_port = 80               // port #
  ip_filtering = []                 // list of whitelisted IPs
  platform = "alb"                  // or "elb" 
}

#######################################################
# DATABASES
#######################################################

resource "aptible_db" "<db_handle" {
  env_id = "<env_id>"
  handle = "<db_handle>"
  db_type = "<db_type>"            // E.G. "postgresql", "mongodb", etc.
  container_size = "1024"
  disk_size = "10"
}

#######################################################
# REPLICAS
#######################################################

resource "aptible_replica" "<replica_handle" {
  env_id = "<env_id>"
  db_id = "<db_id>"
  handle = "<replica_handle>"
}
