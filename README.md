# Aptible Terraform Provider 

To create an app:
- Create a file named `main.tf`
- Add your app's metadata. 
    -   You can see an example in `examples/main.tf`

Whenever a change is made:
- Build the plugin: `go build -o terraform-provider-aptible`
- Initialize the plugin:   `terraform init`
- See what changes will be made: `terraform plan` 
- Apply the changes: `terraform apply`

