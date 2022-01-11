terraform {
  required_providers {
    tf-secrets-to-k8s-sops = {
      source = "argyle-engineering/k8s-sops-secrets"
    }
  }
}

provider "tf-secrets-to-k8s-sops" {
  sops_config = file("${path.module}/.sops.yaml")
}

resource "sops_secret" "example" {
  provider = tf-secrets-to-k8s-sops
  unencrypted_text = file("example.json")
  namespace = "default"
  name = "example"
}

output "encrypted_secret" {
  value = sops_secret.example.encrypted_text
}