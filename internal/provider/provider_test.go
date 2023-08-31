// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the adcs client is properly configured.
	// It is also possible to use the ADCS_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "microsoftadcs" {
	use_ntlm = true
}
`
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"microsoftadcs": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if v := os.Getenv("ADCS_HOST"); v == "" {
		t.Fatal("ADCS_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("ADCS_USERNAME"); v == "" {
		t.Fatal("ADCS_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("ADCS_PASSWORD"); v == "" {
		t.Fatal("ADCS_PASSWORD must be set for acceptance tests")
	}
}
