

provider "scaffolding" {
  # example configuration here
}

resource "scaffolding_resource" "example" {
  sample_attribute = "foo"
}

data "scaffolding_data_source" "example" {
  sample_attribute = "foo"
}