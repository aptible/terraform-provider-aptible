# Aptible Terraform Provider

The provider is available on the [Terraform Registry](https://registry.terraform.io/providers/aptible/aptible/latest) and can be installed using the normal Terraform [configuration process](https://www.terraform.io/docs/language/providers/configuration.html).

To create an app:

- Create a file named `main.tf`
- Add your app's metadata.
  - You can see an example in `examples/demo.tf`

## Developing the provider

Whenever a change is made:

- Install the plugin locally: `make local-install`
- Initialize the plugin: `terraform init`
  - Your old provider lockfile may need to be removed: `rm .terraform.lock.hcl`
- See what changes will be made: `terraform plan`
- Apply the changes: `terraform apply`

### Testing with an unreleased version of aptible-api-go

To switch to an unreleased version of `aptible-api-go`, run:

```shell
go get github.com/aptible/aptible-api-go@COMMIT
go mod vendor
```

replacing `COMMIT` with the commit you want to test. The specified version will
be pulled and you can start testing with it. A branch or tag can be used instead
of a commit, however, if the branch or tag is updated, go will cache the
download and subsequent `go get` commands will not update the package.

When testing is complete and the new version of the package is release, `go.mod`
can be updated with the desired version, then:

```shell
go mod tidy
go mod vendor
```

will pull the correct package version and update the vendored packages.

## Manual Installation

If you are using a Terraform version that cannot install the provider from the registry,
then you may attempt a local installation. However, we do not test this process and cannot
ensure it works.

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
