package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//"github.com/google/go-github/v41/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	client := meta.(*apiClient)

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	// create a temporary directory to run sops command from
	// this is required since there is no pragmatic way to send .sops.yaml to the sops binary
	tmpDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		return diag.Errorf("failed to create tmp sops dir: %s", err)
	}

	// remove the dir after apply is done
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	// write out our .sops.yaml to our tmp dir
	f, err := os.Create(fmt.Sprintf("%s/.sops.yaml", tmpDir))
	if err != nil {
		return diag.Errorf("failed to create .sops.yaml file: %s", err)
	}
	_, err = f.WriteString(client.SopsConfig)
	if err != nil {
		return diag.Errorf("failed to write .sops.yaml file to sops dir: %s", err)
	}

	// install kubectl if it does not exist
	if err = Exists("kubectl"); err != nil {
		kubectl := Kubectl{}
		err = kubectl.install()
		if err != nil {
			return diag.Errorf("could not install kubectl: %s", err)
		}
	}

	// install sops if it does not exist
	if err = Exists("sops"); err != nil {
		sops := SOPS{}
		err = sops.install()
		if err != nil {
			return diag.Errorf("could not install sops: %s", err)
		}
	}

	// test for bash
	if err = Exists("bash"); err != nil {
		return diag.Errorf("bash is required to run this provider")
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

	out, err := ExecuteBash(fmt.Sprintf("echo '%s' | sops -e /dev/stdin", kubeSecret.String()), tmpDir)

	if err != nil {
		return diag.Errorf("error while creating sops encrypted kubernetes secret: %s", err)
	}

	log.Println(out) // successful SOPS generated secret
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

	//return diag.Errorf("not implemented")
	return nil
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
