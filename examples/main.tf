terraform {
  required_providers {
    tf-secrets-to-k8s-sops = {
      source = "argyle/tf-secrets-to-k8s-sops"
    }
  }
  required_version = "~> 1.0.3"
}

provider "tf-secrets-to-k8s-sops" {
  gh_token = "something something bobs uncle"
}

resource "sops_github_secret" "example" {
  provider = tf-secrets-to-k8s-sops
  value = "super secret value"
}