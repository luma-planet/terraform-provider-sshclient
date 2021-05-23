# Terraform Provider SSH Client

[![](https://img.shields.io/github/go-mod/go-version/luma-planet/terraform-provider-sshclient?style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient)
[![Go Report Card](https://goreportcard.com/badge/github.com/luma-planet/terraform-provider-sshclient)](https://goreportcard.com/report/github.com/luma-planet/terraform-provider-sshclient)
[![](https://img.shields.io/github/workflow/status/luma-planet/terraform-provider-sshclient/test?label=test&style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient/actions/workflows/test.yml)
[![](https://img.shields.io/github/workflow/status/luma-planet/terraform-provider-sshclient/staticcheck?label=staticcheck&style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient/actions/workflows/staticcheck.yml)
[![](https://img.shields.io/github/workflow/status/luma-planet/terraform-provider-sshclient/fmt?label=fmt&style=flat-square)](https://github.com/luma-planet/terraform-provider-sshclient/actions/workflows/fmt.yml)

## Installation

```
terraform {
  required_providers {
    sshclient = {
      source  = "luma-planet/sshclient"
      version = "1.0"
    }
  }
}
```

## Development

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
