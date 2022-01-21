package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

type Branch struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Sha  string `json:"sha"`
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"object"`
}

func createBranch(accessToken string, fromBranch string) error {

	branches, err := getBranches(accessToken)

	if err != nil {
		return err
	}

	for _, branch := range branches {
		fmt.Println(branch.Ref)
	}

	return nil
}

func getBranches(accessToken string) ([]Branch, error) {

	org := "argyle-systems"
	repo := "gh-actions-playground"

	client := createGHClient(accessToken)

	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads", org, repo))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad request")
	}

	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var branches []Branch

	err = json.Unmarshal(bodyBytes, &branches)
	if err != nil {
		return nil, err
	}

	return branches, err
}

func createGHClient(accessToken string) *http.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return tc
}
