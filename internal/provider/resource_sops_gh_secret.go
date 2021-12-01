package provider

import (
	"context"
	"github.com/google/go-github/v41/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/oauth2"
	"log"
)

func resourceSopsGithubSecret() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Sample resource in the Terraform provider SopsGithubSecret.",

		CreateContext: resourceSopsGithubSecretCreate,
		ReadContext:   resourceSopsGithubSecretRead,
		UpdateContext: resourceSopsGithubSecretUpdate,
		DeleteContext: resourceSopsGithubSecretDelete,

		Schema: map[string]*schema.Schema{
			"value": {
				// This description is used by the documentation generator and the language server.
				Description: "You're secret string value",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
		},
	}
}

func resourceSopsGithubSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: client.GhToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)

	// list all repositories for the authenticated user
	repos, _, err := ghClient.Repositories.List(ctx, "", nil)

	if err != nil {
		return diag.Errorf("error while accessing Github: %s", err)
	}

	log.Printf("%v", repos)

	return diag.Errorf("not implemented")
}

func resourceSopsGithubSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceSopsGithubSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceSopsGithubSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}
