#######################################################
# PROVIDER
#######################################################

# terraform {
#   required_providers {
#     aptible = {
#       source  = "aptible/aptible"
#       version = "~> 0.1"
#     }
#   }
# }

#######################################################
# ENVIRONMENTS
#######################################################

# data "aptible_environment" "example" {
#   handle = "example"
# }

#######################################################
# APPS
#######################################################

# resource "aptible_app" "nginx" {
#   env_id = data.aptible_environment.demo.env_id
#   handle = "nginx"
#   config = {
#     "APTIBLE_DOCKER_IMAGE" = "nginx"
#   }
# }

#######################################################
# ENDPOINTS
#######################################################

# resource "aptible_endpoint" "https" {
#   env_id         = data.aptible_environment.demo.env_id
#   resource_id    = aptible_app.nginx.id
#   resource_type  = "app"   // other options: database
#   process_type   = "cmd"   // cmd is the default process_type
#   endpoint_type  = "https" // other options: tcp, tls
#   internal       = true    // or false for external
#   container_port = 80      // port #
#   platform       = "alb"   // or "elb"
# }
