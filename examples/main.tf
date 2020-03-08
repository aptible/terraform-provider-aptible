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
<<<<<<< HEAD
=======
  type = "HTTPS"                    // other options: TCP, TLS
>>>>>>> 66f699118b8596b522aaf69c0a2fdd8d42915726
  internal = true                   // or false for external
  container_port = 80               // port #
  ip_filtering = []                 // list of whitelisted IPs
  platform = "alb"                  // or "elb" 
}
