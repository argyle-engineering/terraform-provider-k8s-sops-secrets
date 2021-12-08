package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"os"
	"strings"
)

func resourceSopsSecret() *schema.Resource {
	return &schema.Resource{
		Description: "A SOPs encrypted Kubernetes Secret",

		CreateContext: resourceSopsSecretCreate,
		ReadContext:   resourceSopsSecretRead,
		UpdateContext: resourceSopsSecretUpdate,
		DeleteContext: resourceSopsSecretDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Secret name",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
			"unencrypted_text": {
				Description: "Unencrypted string value",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
			"encrypted_text": {
				Description: "Encrypted string value",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    false,
				Computed:    true,
			},
			"namespace": {
				Description: "Kubernetes namespace where you want your secret to exist",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
		},
	}
}

func resourceSopsSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// using the meta value to retrieve our client from the provider configure method
	client := meta.(*apiClient)

	d.SetId(fmt.Sprintf("%s-%s", d.Get("name"), d.Get("namespace")))

	// ============================== Generate SOPs encrypted K8s Secret ===============================================

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

	if _, err = f.WriteString(client.SopsConfig); err != nil {
		return diag.Errorf("failed to write .sops.yaml file to sops dir: %s", err)
	}

	var missingDependencies []string

	// test for kubectl
	if err = Exists("kubectl"); err != nil {
		missingDependencies = append(missingDependencies, "kubectl")
	}

	// test for sops
	if err = Exists("sops"); err != nil {
		missingDependencies = append(missingDependencies, "sops")
	}

	// test for bash
	if err = Exists("bash"); err != nil {
		missingDependencies = append(missingDependencies, "bash")
	}

	if len(missingDependencies) > 0 {
		return diag.Errorf("the following dependencies are required to run this provider: %s", strings.Join(missingDependencies, ", "))
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

	sopsSecret, err := ExecuteBash(fmt.Sprintf("echo '%s' | sops -e /dev/stdin", kubeSecret.String()), tmpDir)

	if err != nil {
		return diag.Errorf("error while creating sops encrypted kubernetes secret: %s", err)
	}

	if err = d.Set("encrypted_text", sopsSecret); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSopsSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceSopsSecretCreate(ctx, d, meta)
}

func resourceSopsSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceSopsSecretCreate(ctx, d, meta)
}

func resourceSopsSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
