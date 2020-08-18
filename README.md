# Aptible Terraform Provider

To create an app:

- Create a file named `main.tf`
- Add your app's metadata.
  - You can see an example in `examples/main.tf`

Whenever a change is made:

- Build the plugin: `go build -o terraform-provider-aptible`
- Initialize the plugin: `terraform init`
- See what changes will be made: `terraform plan`
- Apply the changes: `terraform apply`

### Verifying the Releases

All of the precompiled binaries available on the release page have checksums published to
verify the integrity of the zip archives. To verify the checksums, we have signed them with a
GPG key.

The public key is available at ...

Once you have the public key, you can use it to verify the checksums and then, in turn, use
those to verify the binaries. For example:

```
gpg --import aptible_terraform_provider.pub
gpg --verify terraform-provider-aptible_${VERSION}_SHA256SUMS.sig
sha256sum -c --ignore-missing terraform-provider-aptible_0.1_SHA256SUMS
```

The exact commands may vary on different systems.
