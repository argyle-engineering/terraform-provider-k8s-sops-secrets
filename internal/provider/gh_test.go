package provider

import (
	"os"
	"testing"
)

func TestCreateBranch(t *testing.T) {

	token := os.Getenv("GITHUB_TOKEN")

	err := createBranch(token, "master")
	if err != nil {
		t.Error(err)
	}
}

func TestCreateBranchFromNonExistingBranch(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	err := createBranch(token, "masterrrrr")
	if err == nil {
		t.Error("branch was created from non-existent branch")
	}
}
