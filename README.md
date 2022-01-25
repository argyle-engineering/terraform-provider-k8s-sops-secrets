# Terraform Provider Terraform Secrets to Kubernetes SOPS

This repository is a *provider* for a [Terraform](https://www.terraform.io) responsible for adding Terraform Secret values into Kubernetes based (GitOps) Github repo.

## Repository Layout
 - A resource, and a data source (`internal/provider/`),
 - Examples (`examples/`) and generated documentation (`docs/`),
 - Miscellaneous meta files.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.15

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command: 
```sh
$ go install
```

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `/usrlocal//bin` directory.

Add the following to your `~/.terraformrc` to use the development version instead of pulling from the remote registry.
```yaml
provider_installation {

  dev_overrides {
      "argyle-engineering/k8s-sops-secrets" = "/usr/local/bin/"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

set `TF_LOG` environment variable to `TRACE` to view all logs and debug information. 
```shell
export TF_LOG=TRACE
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

## Release instructions

In order to release using [gorelaser](https://goreleaser.com/quick-start/) you will need a GitHub Personal Access Token with a full repo scope, you will need
a valid GPG key as well.
```sh
export GITHUB_TOKEN=<TOKEN>
export GPG_TTY=$(tty)
export GPG_FINGERPRINT=<FINGERPRINT>
```
