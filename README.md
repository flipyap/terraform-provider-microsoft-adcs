# Terraform Provider Microsoft ADCS (Terraform Plugin Framework)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Example of using the provider below:

```hcl
terraform {
  required_providers {
    microsoftadcs = {
      source = "registry.terraform.io/flipyap/microsoft-adcs"
    }
  }
}

provider "microsoftadcs" {
  host = "server.company.local"
  username = "Username"
}

data "microsoftadcs_certificate" "example" {
  id = "525135"
}

output "example_certificate" {
  value = data.microsoftadcs_certificate.example
}

resource "microsoftadcs_certificate" "my_cert" {
  certificate_signing_request = base64decode(local.csr)
  template = "User"

}

output "my_cert_certs" {
  value = microsoftadcs_certificate.my_cert
}
```


## Authenticating

Curerently this provider only supports kerberos auth, but there is plans to add NTLM basic auth as well.

To auth you can provider the `krb5conf` attribute to the provider or set the `ADCS_KRB5CONF` environment variable. Lastly, the client will look at `/etc/krb5.conf` for configuration.

>	&#10071; **Known Issues**
- Client suppports limited amount of templates. Should opent that up.
- You need to indicate KDC servers in your krb5conf or include the `dns_lookup_kdc = true` key


## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
