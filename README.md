# Terraform Provider SSH Client

[![](https://img.shields.io/github/go-mod/go-version/luma-planet/terraform-provider-sshclient?style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient)
[![](https://img.shields.io/github/workflow/status/luma-planet/terraform-provider-sshclient/test?label=test&style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient/actions/workflows/test.yml)
[![](https://img.shields.io/github/workflow/status/luma-planet/terraform-provider-sshclient/staticcheck?label=staticcheck&style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient/actions/workflows/staticcheck.yml)

Run the following command to build the provider

```shell
go build -o terraform-provider-sshclient
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```

## TODO for v1

- [x] release
- [x] unit test
- [x] acc test
- [x] CI
- [ ] badges
- [ ] doc
