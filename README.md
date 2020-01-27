# Aptible Terraform Provider 

Whenever a change is made:
- Build the plugin: `go build -o terraform-provider-aptible`
- Initialize the plugin:   `terraform init`

To create an app:
- Create a file named `main.tf`
- Add the app's metadata:
```
resource "aptible_app" "<name_of_app>" {
    account_id = "<your_account_id>"
    handle = "<name_of_app>"
    env = {
        "env_var name" = "env_var value",
        ...
    }
}
```

To see what changes will be made: `terraform plan` 

To apply the changes: `terraform apply`

