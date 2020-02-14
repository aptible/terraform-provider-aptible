// Examples of resources

resource "aptible_app" "<app_handle>" {
    account_id = "<account_id>"
    handle = "<app_handle>"
    env = {
        "APTIBLE_DOCKER_IMAGE" = "<docker_image>"
        "ANOTHER_VAR" = "value"
    }
}


