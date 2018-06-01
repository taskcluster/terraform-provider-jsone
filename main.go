package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/taskcluster/terraform-provider-jsone/jsoneprovider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: jsoneprovider.Provider})
}
