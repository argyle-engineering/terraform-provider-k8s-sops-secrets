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

func resourceSopsSecretCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	d.SetId(getID(d))

	depErr := dependencyChecks()
	if depErr != nil {
		return depErr
	}

	// ============================== Generate SOPs encrypted K8s Secret ===============================================

	sopsSecret, depErr := createSOPSSecret(d, meta)

	if depErr != nil {
		return diag.Errorf("error while creating sops encrypted kubernetes secret: %s", depErr)
	}

	if err := d.Set("encrypted_text", sopsSecret); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSopsSecretRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	d.SetId(getID(d))

	sopsSecret, depErr := createSOPSSecret(d, meta)
	if depErr != nil {
		return depErr
	}

	err := d.Set("state", sopsSecret)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSopsSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceSopsSecretCreate(ctx, d, meta)
}

func resourceSopsSecretDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func dependencyChecks() diag.Diagnostics {
	var missingDependencies []string

	// test for sops
	if err := Exists("sops"); err != nil {
		missingDependencies = append(missingDependencies, "sops")
	}

	// test for bash
	if err := Exists("bash"); err != nil {
		missingDependencies = append(missingDependencies, "bash")
	}

	if len(missingDependencies) > 0 {
		return diag.Errorf("the following dependencies are required to run this provider: %s", strings.Join(missingDependencies, ", "))
	}

	return nil
}

func setupSOPSConfigFile(meta interface{}) (diag.Diagnostics, string) {

	// using the meta value to retrieve our client from the provider configure method
	client := meta.(*apiClient)

	// create a temporary directory to run sops command from
	// this is required since there is no pragmatic way to send .sops.yaml to the sops binary
	tmpDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		return diag.Errorf("failed to create tmp sops dir: %s", err), ""
	}

	// write out our .sops.yaml to our tmp dir
	f, err := os.Create(fmt.Sprintf("%s/.sops.yaml", tmpDir))
	if err != nil {
		return diag.Errorf("failed to create .sops.yaml file: %s", err), tmpDir
	}

	if _, err = f.WriteString(client.SopsConfig); err != nil {
		return diag.Errorf("failed to write .sops.yaml file to sops dir: %s", err), tmpDir
	}

	return nil, tmpDir
}

func createSOPSSecret(d *schema.ResourceData, meta interface{}) (string, diag.Diagnostics) {
	depErr, tmpDir := setupSOPSConfigFile(meta)

	// remove the dir after apply is done
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	if depErr != nil {
		return "", depErr
	}

	// create k8s secret from secret value
	name := fmt.Sprintf("%s", d.Get("name"))
	value := fmt.Sprintf("%s", d.Get("value"))

	s := NewSecret(name)
	sd := StringData{
		name: value,
	}
	s.StringData = sd
	kubeSecret, err := s.Marshall()

	if err != nil {
		return "", diag.Errorf("error while creating kubernetes secret: %s", err)
	}

	sopsSecret, err := ExecuteBash(fmt.Sprintf("echo \"%s\" | sops --output-type=yaml -e /dev/stdin", kubeSecret), tmpDir)

	if err != nil {
		return "", diag.Errorf("%s", err)
	}

	return sopsSecret, nil

}

func getID(d *schema.ResourceData) string {
	if d.Id() != "" {
		return d.Id()
	}

	return fmt.Sprintf("%s-%s", d.Get("name"), d.Get("namespace"))
}
