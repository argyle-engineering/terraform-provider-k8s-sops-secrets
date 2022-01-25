package provider

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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
			"sops_config": {
				Description: "SOPS config file content in yaml mode",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
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
			"is_base64": {
				Description: "Indicates whether we use stringData or ",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Required:    false,
			},
			"unencrypted_hash": {
				Description: "Unencrypted string md5sum value",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    false,
				Computed:    true,
			},
			"encrypted_text": {
				Description: "Encrypted string value",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    false,
				Computed:    true,
			},
			"namespace": {
				Description: "Kubernetes namespace where you want your stringDataSecret to exist",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
		},
	}
}

func resourceSopsSecretCreate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {

	d.SetId(getID(d))

	depErr := dependencyChecks()
	if depErr != nil {
		return depErr
	}

	// ============================== Generate SOPs encrypted K8s Secret ===============================================

	sopsSecret, kubeSecret, depErr := createSOPSSecret(d)

	if depErr != nil {
		return diag.Errorf("error while creating sops encrypted kubernetes stringDataSecret: %s", depErr)
	}

	rawUnencryptedHash := md5.Sum([]byte(kubeSecret))
	unencryptedHash := hex.EncodeToString(rawUnencryptedHash[:])

	if err := d.Set("unencrypted_hash", unencryptedHash); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("encrypted_text", sopsSecret); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSopsSecretRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {

	sopsSecret, kubeSecret, depErr := createSOPSSecret(d)
	if depErr != nil {
		return depErr
	}

	existingHash := fmt.Sprintf("%s", d.Get("unencrypted_hash"))
	rawNewHash := md5.Sum([]byte(kubeSecret))
	newHash := hex.EncodeToString(rawNewHash[:])

	if newHash != existingHash {
		err := d.Set("encrypted_text", sopsSecret)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceSopsSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceSopsSecretCreate(ctx, d, meta)
}

func resourceSopsSecretDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {

	if err := d.Set("unencrypted_hash", nil); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("encrypted_text", nil); err != nil {
		return diag.FromErr(err)
	}

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

func setupSOPSConfigFile(sopsConfig string) (diag.Diagnostics, string) {

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

	if _, err = f.WriteString(sopsConfig); err != nil {
		return diag.Errorf("failed to write .sops.yaml file to sops dir: %s", err), tmpDir
	}

	return nil, tmpDir
}

func createSOPSSecret(d *schema.ResourceData) (string, string, diag.Diagnostics) {

	sopsConfig := fmt.Sprintf("%s", d.Get("sops_config"))

	depErr, tmpDir := setupSOPSConfigFile(sopsConfig)

	// remove the dir after apply is done
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	if depErr != nil {
		return "", "", depErr
	}

	// create k8s secret from secret value
	name := fmt.Sprintf("%s", d.Get("name"))
	value := fmt.Sprintf("%s", d.Get("value"))
	isBase64, _ := strconv.ParseBool(fmt.Sprintf("%s", d.Get("is_base_64")))

	s := NewSecret(name)

	if isBase64 {
		s.Data = Data{
			name: value,
		}
	} else {
		s.StringData = StringData{
			name: value,
		}
	}

	kubeSecret, err := s.Marshall()

	if err != nil {
		log.Printf("[DEBUG] marshalled output: \n %s \n\n\n")
		return "", "", diag.Errorf("error while creating kubernetes secret: %s", err)
	}

	bashScript := fmt.Sprintf("echo \"%s\" | sops --output-type=yaml -e /dev/stdin", kubeSecret)
	log.Printf("[DEBUG] bashScript output: \n %s \n\n\n", bashScript)

	sopsSecret, err := ExecuteBash(bashScript, tmpDir)

	if err != nil {
		return "", "", diag.Errorf("%s", err)
	}

	return sopsSecret, kubeSecret, nil

}

func getID(d *schema.ResourceData) string {
	if d.Id() != "" {
		return d.Id()
	}

	return fmt.Sprintf("%s-%s", d.Get("name"), d.Get("namespace"))
}
