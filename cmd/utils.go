package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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

func NewClientBasedOnToken(organization string, tokenFile string) *GithubClient {
	token, err := GetAPITokenFromFile(tokenFile)

	if err != nil {
		return NewGithubClientNoAuth(organization)
	}

	return NewGithubClient(organization, token)
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

// CreateNewPR creates a new PR from a branch
func (g GithubClient) CreateNewPR(repo string, branch string, user string) (*github.PullRequest, error) {
	input := &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Release %v", branch)),
		Body:  github.String(fmt.Sprintf("Release %v", branch)),
		Head:  github.String(fmt.Sprintf("%v", branch)),
		Base:  github.String("master"),
	}

	pr, _, err := g.client.PullRequests.Create(context.Background(), g.owner, repo, input)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

// GetLatestRelease get the latest release version
func (g GithubClient) GetLatestRelease(repo string) (string, error) {
	fmt.Println(repo)
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
		return time.Time{}, fmt.Errorf("failed to get release tag: %s", err)
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

// fetchLabel fetches the first label of the PR
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

// UpdateReleaseNotes update github release note
func (g GithubClient) CreateNewRelease(repo string, tag string, releaseNotes string) error {
	repoRelease := &github.RepositoryRelease{
		Name:    github.String(tag),
		TagName: github.String(tag),
		Body:    github.String(releaseNotes),
	}

	_, _, err := g.client.Repositories.CreateRelease(context.Background(), g.owner, repo, repoRelease)
	if err != nil {
		return err
	}

	return nil
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	color.Red(fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
