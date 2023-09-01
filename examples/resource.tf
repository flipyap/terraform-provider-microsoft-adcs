terraform {
  required_providers {
    microsoftadcs = {
      source = "registry.terraform.io/flipyap/microsoft-adcs"
    }
  }
}

provider "microsoftadcs" {
  use_ntlm = true
}

resource "microsoftadcs_certificate" "my_cert" {
  certificate_signing_request = base64decode(local.csr)
  template                    = "User"

}

output "my_cert_certs" {
  value = microsoftadcs_certificate.my_cert
}
