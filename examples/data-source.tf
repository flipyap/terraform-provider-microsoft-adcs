//terraform {
//  required_providers {
//    microsoftadcs = {
//      source = "registry.terraform.io/flipyap/microsoftadcs"
//    }
//  }
//}

provider "microsoftadcs" {
}

data "microsoftadcs_certificate" "example" {
  id = "525135"
}

output "example_certificate" {
  value = data.microsoftadcs_certificate.example
}
