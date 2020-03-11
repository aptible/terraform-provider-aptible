// Examples of resources

#######################################################
# APPS
#######################################################

resource "aptible_app" "<app_handle>" {
    account_id = "<account_id>"
    handle = "<app_handle>"
    env = {
        "APTIBLE_DOCKER_IMAGE" = "<docker_image>"
        "ANOTHER_VAR" = "value"
    }
}

#######################################################
# ENDPOINTS
#######################################################

resource "aptible_endpoint" "<endpoint_name>" {
  account_id = "<account_id>"
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
  account_id = "<account_id>"
  handle = "<db_handle>"
  type = "<db_type>"            // E.G. "postgresql", "mongodb", etc.
  container_size = "1024"
  disk_size = "10"
}