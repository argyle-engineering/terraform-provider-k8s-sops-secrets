terraform {
  required_providers {
    tf-secrets-to-k8s-sops = {
      source = "argyle/tf-secrets-to-k8s-sops"
    }
  }
  required_version = "~> 1.0.3"
}

provider "scaffolding" {
  # example configuration here
}

resource "scaffolding_resource" "example" {
  sample_attribute = "foo"
}

data "scaffolding_data_source" "example" {
  sample_attribute = "foo"
}