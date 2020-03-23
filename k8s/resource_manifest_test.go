package k8s

import (
	"fmt"
	"github.com/hashicorp/terraform-provider-template/template"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

var testProvider terraform.ResourceProvider
var testProviders map[string]terraform.ResourceProvider
var testProviderFunc func() terraform.ResourceProvider
var testProviderFactories func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory

func init() {
	testProvider = Provider().(terraform.ResourceProvider)
	testProviders = map[string]terraform.ResourceProvider{
		"k8s":      testProvider,
		"template": template.Provider(),
	}
	testProviderFactories = func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory {
		return map[string]terraform.ResourceProviderFactory{
			":": func() (terraform.ResourceProvider, error) {
				p := Provider()
				*providers = append(*providers, p.(*schema.Provider))
				return p, nil
			},
		}
	}
	testProviderFunc = func() terraform.ResourceProvider { return testProvider }
}
func TestProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("k8s_manifest.tiller", "content"),
				),
			},
		},
	})
}

func testDataSourceConfig() string {
	return fmt.Sprintf(`
provider "k8s" {
  load_config_file = false
  host = "https://34.65.27.127"
  token = "ya29.c.KpYBwwdU9HnRMNDrakkz9DJHfuoGoRTE9LzT6Mu7OtciGRc--BPuy-vGz3Z5FaEpnGYXnP3XdLp3mPj-iSoQhHmC16rWbkrEHM33g2fJej9vii0Jc-iICWyjdBorvP77y-rEFT5RrF_SfdVjx8qNbK33KV5WSICYgvnIJrbEquGq-qrbaFGvollXZ012C5uoi_KigtNZMgyM"
  cluster_ca_certificate = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURDekNDQWZPZ0F3SUJBZ0lRVW9nL3o3WjZrUCtTRGM4d2h1VFlhekFOQmdrcWhraUc5dzBCQVFzRkFEQXYKTVMwd0t3WURWUVFERXlRM1lqQXhPVGMzTWkwM05EWmxMVFEwTUdZdFlXSTNOeTAxWkdNME4yVTBaakZsWlRFdwpIaGNOTWpBd016SXpNREF4TWpVMFdoY05NalV3TXpJeU1ERXhNalUwV2pBdk1TMHdLd1lEVlFRREV5UTNZakF4Ck9UYzNNaTAzTkRabExUUTBNR1l0WVdJM055MDFaR00wTjJVMFpqRmxaVEV3Z2dFaU1BMEdDU3FHU0liM0RRRUIKQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUMySk5xclczc3ROWFlLZ0xmVVlTcktJcmhoQkV1ZWEvdEp4S0I3UXJzbApkeEppa0Nub090SHVETkFRc1RFUEZqNVZkbkkwaG8wT0ZhaTdCbThhRFpjajNvMXZHb05NUjNCTFpiSU5FbTVRCmxZSmJnbCtDQ3ptdmpTMWY1K0ZVZ0ozZjBFZDhielBJdnlwTERZWGtoSUkwSllCTC95TFJtcWtyZlpKSGpSNE4KTTFlMWhKVC9lcFRFVmlWcUJ2SGhldjRoY2pMYjBBVjBlVFpGejF3UG5YYlF2MWp2VkVCd2RET3VJYjdMU1hrWQpIMzNhYnhLUHNiZ2VYZ04xRlFYNkNlUjIxQ1hPSmVUMHJBVlBjVGpta2h0WWpOSU9rZHljWHlKRFhZT0FoTnZ4ClRUekxFV0Jub2dsdDltUTlzS2NPUUYwekdNZDNDQ1o1RWZYU3g5OUh0OUpkQWdNQkFBR2pJekFoTUE0R0ExVWQKRHdFQi93UUVBd0lDQkRBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDdgpVM1AzY3M4RDJkU3hEVjE4MFkxL3ZxbHBnL3hkdWdmbzFGd3VpN0tOUGM2bzhZaWVsSUx0WDViY2pBK1pyMzQrCnVkZjZuVEtMY0dNeVlFN1lGS1pScStpU0pYb2NLc3BFYVJPeVJzQjJXZjlHYmhvOGVaOGdoUkdvWncyQWwwV1YKQnByUXJBZFdaUmpDQWNaMy8wcVpoNHJpZkFTQmh2ZlNKNElMMlRJNjNyVnVpMGRrd3JxOGlxZno2NU91RU9BOAoxclNMMFQ3MjBJcEdVaFNyakRjZUlUeWZiU2c4dGUwM1g0bHpZamxJTmNSeEs2VlZYNjYvUWF2NWxUZDN4elFrClZGWVFhc1J3R2hxaFlUdUM5T3hNNUhLVG43cklpd2JDc1Z3VGcrWFNNb29hL3lSMHNyT2tZUTBDRFRlM0pORXQKdHpGbE5MTHhZSWYyMDVyaUcrZ3IKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
}
data "template_file" "tiller_template" {
  template = "${file("../examples/manifests/tiller.yaml")}"
}
resource "k8s_manifest" "tiller" {
  content = "${data.template_file.tiller_template.rendered}"
}
`)
}
