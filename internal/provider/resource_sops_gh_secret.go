package provider

import (
	"context"
	"fmt"
	//"github.com/google/go-github/v41/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	//"golang.org/x/oauth2"
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
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Secret name",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
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
	// client := meta.(*apiClient)

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	// install kubectl if it does not exist
	if err := Exists("kubectl"); err != nil {
		kubectl := Kubectl{}
		err = kubectl.install()
		if err != nil {
			return diag.Errorf("could not install kubectl: %s", err)
		}
	}

	// create k8s secret from secret value
	err, kubeSecret := LocalExecutor(
		"kubectl",
		"create",
		"secret",
		"generic",
		fmt.Sprintf("%s", d.Get("name")),
		fmt.Sprintf("--from-literal=%s='%s'", d.Get("name"), d.Get("value")),
		"--namespace",
		"monitoring",
		"--dry-run=client",
		"-o",
		"yaml",
	)

	if err != nil {
		return diag.Errorf("error while creating kubernetes secret: %s", err)
	}

	// TODO: Encrypt string with SOPs
	log.Println(kubeSecret.String())

	// ======================================= Add file to GH ========================================================

	//ts := oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: client.GhToken},
	//)
	//tc := oauth2.NewClient(ctx, ts)
	//
	//ghClient := github.NewClient(tc)
	//
	//// list all repositories for the authenticated user
	//repos, _, err := ghClient.Repositories.List(ctx, "", nil)
	//
	//if err != nil {
	//	return diag.Errorf("error while accessing Github: %s", err)
	//}
	//
	//log.Printf("%v", repos)

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
