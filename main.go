package main

import (
	"github.com/greg-solutions/terraform-provider-k8s/k8s"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: k8s.Provider,
	})
}
