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
data "template_file" "tiller_template" {
  template = "${file("../examples/manifests/tiller.yaml")}"
}

resource "k8s_manifest" "tiller" {
  content = "${data.template_file.tiller_template.rendered}"
}
`)
}
