package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"strings"
	"time"
)

type PR struct {
	Title string
	Link  string
	Type  string
}

type ByLabel []PR

func (l ByLabel) Len() int {
	return len(l)
}

func (l ByLabel) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l ByLabel) Less(i, j int) bool {
	return l[i].Type < l[j].Type
}

// GithubClient masks RPCs to github as local procedures
type GithubClient struct {
	client *github.Client
	owner  string
	token  string
}

// NewGithubClient creates a new GithubClient with proper authentication
func NewGithubClient(owner, token string) *GithubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return &GithubClient{client, owner, token}
}

// NewGithubClientNoAuth creates a new GithubClient without authentication
// useful when only making GET requests
func NewGithubClientNoAuth(owner string) *GithubClient {
	client := github.NewClient(nil)
	return &GithubClient{client, owner, ""}
}

// GetAPITokenFromFile returns the github api token from tokenFile
func GetAPITokenFromFile(tokenFile string) (string, error) {
	b, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(b[:]))
	return token, nil
}

// GetLatestRelease get the latest release version
func (g GithubClient) GetLatestRelease(repo string) (string, error) {
	release, _, err := g.client.Repositories.GetLatestRelease(context.Background(), g.owner, repo)
	if err != nil {
		return "", err
	}
	return *release.TagName, nil
}

// GetReleaseTagCreationTime gets the creation time of a lightweight tag created by release
func (g GithubClient) GetReleaseTagCreationTime(repo, tag string) (time.Time, error) {
	release, _, err := g.client.Repositories.GetReleaseByTag(context.Background(), g.owner, repo, tag)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to to get release tag: %s", err)
	}
	return release.GetCreatedAt().Time, nil
}

// SearchIssues get issues/prs based on query
func (g GithubClient) SearchIssues(queries []string, keyWord string) (*github.IssuesSearchResult, error) {
	q := strings.Join(queries, " ")

	issueResult, _, err := g.client.Search.Issues(context.Background(), q, nil)
	if err != nil {
		fmt.Printf("Failed to search issues")
		return nil, err
	}

	return issueResult, nil
}

func fetchLabel(labels []github.Label) string {
	for _, label := range labels {
		if *label.Name != "release-note" {
			return *label.Name
		}
	}
	return "Other"
}

// ContainsString finds if target presents in the given slice
func ContainsString(slice []string, target string) bool {
	for _, element := range slice {
		if element == target {
			return true
		}
	}
	return false
}

func (g GithubClient) UpdateReleaseNotes(repo string, tag string, releaseNotes string) error {
	release, _, err := g.client.Repositories.GetReleaseByTag(context.Background(), g.owner, repo, tag)
	if err != nil {
		return err
	}

	*release.Body = releaseNotes
	_, _, err = g.client.Repositories.EditRelease(context.Background(), g.owner, repo, *release.ID, release)
	if err != nil {
		return err
	}

	return nil
}
