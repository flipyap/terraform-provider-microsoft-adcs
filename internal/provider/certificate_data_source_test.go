package provider

//import (
//	"testing"
//
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//)
//
//func TestAccCertificateDataSource(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		Steps: []resource.TestStep{
//			// Read testing
//			{
//				Config: providerConfig + `data "microsoftadcs_certificate" "test" {
//						id = "525135"
//						}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					// Verify number of coffees returned
//					resource.TestCheckResourceAttr("data.microsoftadcs_certificate.test", "id", "525135"),
//					// Verify the first coffee to ensure all attributes are set
//					resource.TestCheckResourceAttrSet("data.microsoftadcs_certificate.test", "certificate_b64"),
//					resource.TestCheckResourceAttrSet("data.microsoftadcs_certificate.test", "certificate_chain_b64"),
//				),
//			},
//		},
//	})
//}
