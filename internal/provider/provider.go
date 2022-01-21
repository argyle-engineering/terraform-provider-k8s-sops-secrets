package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			ResourcesMap: map[string]*schema.Resource{
				"sops_secret": resourceSopsSecret(),
			},
			Schema: map[string]*schema.Schema{},
		}

		p.ConfigureContextFunc = configure(p)

		return p
	}
}

type apiClient struct{}

func configure(_ *schema.Provider) func(_ context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(_ context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &apiClient{}, nil
	}
}
