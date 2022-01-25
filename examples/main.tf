terraform {
  required_providers {
    tf-secrets-to-k8s-sops = {
      source = "argyle-engineering/k8s-sops-secrets"
      version = "1.7.0"
    }
  }
}

provider "tf-secrets-to-k8s-sops" {}

resource "sops_secret" "example" {
  provider = tf-secrets-to-k8s-sops
  unencrypted_text = file("example.json")
  namespace = "default"
  name = "example"
  is_base64 = false
  sops_config = file("${path.module}/.sops.yaml")
}

output "encrypted_secret" {
  value = sops_secret.example.encrypted_text
}