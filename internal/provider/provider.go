package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
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
				"sops_github_secret": resourceSopsGithubSecret(),
			},
			Schema: map[string]*schema.Schema{
				"gh_token": {
					Sensitive: true,
					Type:      schema.TypeString,
					Optional:  true,
				},
				"repo": {
					Sensitive: false,
					Type:      schema.TypeString,
					Optional:  true,
				},
				"sops_config": {
					Sensitive: true,
					Type:      schema.TypeString,
					Optional:  true,
				},
			},
		}

		p.ConfigureContextFunc = configure(p)

		return p
	}
}

type apiClient struct {
	GhToken    string
	Repo       string
	SopsConfig string
}

func configure(_ *schema.Provider) func(_ context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(_ context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {

		requiredValues := []string{"gh_token", "repo", "sops_config"}

		values, err := validate(requiredValues, rd)

		if err != nil {
			return nil, err
		}

		return &apiClient{
			GhToken:    fmt.Sprintf("%s", values["gh_token"]),
			Repo:       fmt.Sprintf("%s", values["repo"]),
			SopsConfig: fmt.Sprintf("%s", values["sops_config"]),
		}, nil
	}
}

func validate(requiredValues []string, rd *schema.ResourceData) (map[string]interface{}, diag.Diagnostics) {
	var missing []string

	found := make(map[string]interface{})

	for _, value := range requiredValues {
		v, _ := rd.GetOk(value)

		if v == "" {
			missing = append(missing, fmt.Sprintf("'%s'", value))
		} else {
			found[value] = v
		}
	}

	if len(missing) > 0 {
		return found, diag.Errorf("missing required configuration(s): %s ", strings.Join(missing, ","))
	}

	return found, nil
}
