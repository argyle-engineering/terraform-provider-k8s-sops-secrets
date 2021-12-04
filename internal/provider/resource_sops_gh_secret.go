package provider

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v41/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSopsGithubSecret() *schema.Resource {
	return &schema.Resource{
		Description: "A Github based SOPs encrypted Kubernetes Secret",

		CreateContext: resourceSopsGithubSecretCreate,
		ReadContext:   resourceSopsGithubSecretRead,
		UpdateContext: resourceSopsGithubSecretUpdate,
		DeleteContext: resourceSopsGithubSecretDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Secret name",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
			"value": {
				Description: "You're secret string value",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
			"namespace": {
				Description: "namespace to create secret",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
			"base_branch": {
				Description: "git branch where changes should be merged into",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
			},
		},
	}
}

func resourceSopsGithubSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	sopsSecret, err := ExecuteBash(fmt.Sprintf("echo '%s' | sops -e /dev/stdin", kubeSecret.String()), tmpDir)

	if err != nil {
		return diag.Errorf("error while creating sops encrypted kubernetes secret: %s", err)
	}

	// ======================================= Add file to our git repo =================================================

	// create a temporary directory to run our Github Repo
	repoDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		return diag.Errorf("failed to create temp repo dir: %s", err)
	}

	// remove the dir after apply is done
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(repoDir)

	// clone repo
	repoName := fmt.Sprintf("https://%s@github.com/%s.git", client.GhToken, client.Repo)
	repo, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL: repoName,
	})
	if err != nil {
		return diag.Errorf("error while cloning repo '%s': %s", repoName, err)
	}

	branchName := plumbing.NewBranchReferenceName(fmt.Sprintf("%s-secret-from-terraform", d.Get("name")))
	headRef, err := repo.Head()
	if err != nil {
		return diag.Errorf("error while getting head reference", err)
	}

	ref := plumbing.NewHashReference(branchName, headRef.Hash())
	if err = repo.Storer.SetReference(ref); err != nil {
		return diag.Errorf("error while setting reference", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return diag.Errorf("error while getting working tree: %s", err)
	}

	if err = worktree.Checkout(&git.CheckoutOptions{Branch: ref.Name()}); err != nil {
		return diag.Errorf("error while checking out new branch %s", err)
	}

	// add files and commit
	fileName := fmt.Sprintf("%s.enc.yaml", d.Get("name"))
	fullFileName := filepath.Join(repoDir, fileName)
	if err = ioutil.WriteFile(fullFileName, []byte(sopsSecret), 0644); err != nil {
		return diag.Errorf("error while writing sops file: %s", err)
	}

	// Adds the new file to the staging area.
	if _, err = worktree.Add(fileName); err != nil {
		return diag.Errorf("error while adding file to working tree: %s - %s", err, fileName)
	}

	// commit
	_, err = worktree.Commit(fmt.Sprintf("adding terraform '%s' secret", d.Get("name")), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "github-actions[bot]",
			Email: "41898282+github-actions[bot]@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return diag.Errorf("error while committing to working tree: %s", err)
	}

	if err = repo.Push(&git.PushOptions{}); err != nil {
		return diag.Errorf("error while pushing to remote: %s", err)
	}

	// ======================================= Create GH PR  ===========================================================

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: client.GhToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)

	repoStr := strings.Split(client.Repo, "/")

	// create PR Request
	_, _, err = ghClient.PullRequests.Create(ctx, repoStr[0], repoStr[1], &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Add terraform '%s' secret", d.Get("name"))),
		Body:  github.String("Resource created via Terraform :robot:"),
		Base:  github.String(fmt.Sprintf("%s", d.Get("base_branch"))),
		Head:  github.String(string(branchName)),
		Draft: github.Bool(false),
	})

	if err != nil {
		return diag.Errorf("error while creating Github Pull Request: %s", err)
	}

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
