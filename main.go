package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider"
)

func main() {
	opts := &plugin.ServeOpts{ProviderFunc: provider.New()}
	plugin.Serve(opts)
}
