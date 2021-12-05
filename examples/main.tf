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
  repo = "argyle-systems/gh-actions-playground"
  sops_config = file("${path.module}/.sops.yaml")
}

resource "sops_github_secret" "example" {
  provider = tf-secrets-to-k8s-sops
  value = "super secret value"
  namespace = "default"
  name = "example"
  base_branch = "master"
  remote_dir = "secrets/"
}

output "PR_URL" {
  value = sops_github_secret.example.pr_url
}
