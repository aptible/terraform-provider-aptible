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

The public key ID is `0xa1b845b9417ca47a02dd7457fb0996ce6372f7ad` and it is available at [the SKS server pool](http://keyserver.ubuntu.com/pks/lookup?op=get&search=0xa1b845b9417ca47a02dd7457fb0996ce6372f7ad)

Once you have the public key, you can use it to verify the checksums and then, in turn, use
those to verify the binaries. For example:

```
gpg --keyserver keyserver.ubuntu.com --recv-key 0xa1b845b9417ca47a02dd7457fb0996ce6372f7ad
gpg --verify terraform-provider-aptible_${VERSION}_SHA256SUMS.sig
sha256sum -c --ignore-missing terraform-provider-aptible_0.1_SHA256SUMS
```

The exact commands may vary on different systems.
