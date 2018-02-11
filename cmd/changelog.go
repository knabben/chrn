package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

// changelogCmd represents the changelog command
var (
	label           string
	currentRelease  string
	previousRelease string

	changelogCmd = &cobra.Command{
		Use:   "changelog",
		Short: "Fetch last changelog from branches",
		Long:  ``,
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			gh := NewClientBasedOnToken(org, token)

			color.Cyan(">>> Start fetching unreleased release note from %s/%s", org, repo)
			query, err := createQueryString(repo, gh)
			CheckIfError(err)

			color.Cyan(">>> Getting PRs for %v", query)
			prList, err := gh.SearchIssues(query, "")
			CheckIfError(err)

			color.Cyan(">>> Modifying changelog file")
			content := groupOtherLabels(prList)
			fmt.Println(content)

			chFile, err := os.Open(file)
			CheckIfError(err)
			defer chFile.Close()

			newLines := ReadFileAndReplace2(chFile, content)
			WriteLinesNewFile(newLines, file)
		},
	}
)

// ReadFileAndReplace read a changelog and replace for new entries
func ReadFileAndReplace2(chFile *os.File, content string) []string {
	scanner := bufio.NewScanner(chFile)
	scanner.Split(bufio.ScanLines)

	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "## [Unreleased]") {
			// Title header change.
			line = fmt.Sprintf("%v%v", line, content)
		}
		lines = append(lines, fmt.Sprintf("%v\n", line))
	}
	return lines
}

func createQueryString(repo string, gh *GithubClient) ([]string, error) {
	var queries []string

	color.Green(fmt.Sprintf("Last release version: %s", currentRelease))

	currentRelease, err := gh.GetLatestRelease(repo)
	CheckIfError(err)

	startTime, err := getReleaseTime(repo, currentRelease, gh)
	CheckIfError(err)

	endTime := time.Now().UTC().Format(time.RFC3339)

	queries = addQuery(queries, "repo", org, "/", repo)
	queries = addQuery(queries, "label", label)
	queries = addQuery(queries, "is", "merged")
	queries = addQuery(queries, "type", "pr")
	queries = addQuery(queries, "base", "master")
	queries = addQuery(queries, "merged", startTime, "..", endTime)

	return queries, nil
}

func addQuery(queries []string, queryParts ...string) []string {
	if len(queryParts) < 2 {
		log.Printf("Not enough to form a query: %v", queryParts)
		return queries
	}
	for _, part := range queryParts {
		if part == "" {
			return queries
		}
	}

	return append(queries, fmt.Sprintf("%s:%s", queryParts[0], strings.Join(queryParts[1:], "")))
}

func getReleaseTime(repo, release string, gh *GithubClient) (string, error) {
	time, err := getReleaseTagCreationTime(repo, release, gh)
	if err != nil {
		log.Println("Failed to get created time of this release tag")
		return "", err
	}
	t := time.UTC()
	timeString := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return timeString, nil
}

func getReleaseTagCreationTime(repo, tag string, gh *GithubClient) (createTime time.Time, err error) {
	createTime, err = gh.GetReleaseTagCreationTime(repo, tag)
	if err != nil {
		log.Printf("Cannot get the creation time of %s/%s", repo, tag)
		return time.Time{}, err
	}
	return createTime, nil
}

func groupOtherLabels(issuesResult *github.IssuesSearchResult) string {
	prGrouper := []PR{}
	existentLabels := make([]string, 3)

	for _, issue := range issuesResult.Issues {
		prGrouper = append(
			prGrouper, PR{
				Title: *issue.Title,
				Link:  *issue.URL,
				Type:  fetchLabel(issue.Labels),
			},
		)
	}
	sort.Sort(ByLabel(prGrouper))

	content := ""
	for _, issue := range prGrouper {
		if !ContainsString(existentLabels, issue.Type) {
			content += fmt.Sprintf("\n## %s\n", strings.Title(issue.Type))
			existentLabels = append(existentLabels, issue.Type)
		}
		content += fmt.Sprintf("- %s\n", issue.Title)
	}
	return content
}

func init() {
	rootCmd.AddCommand(changelogCmd)

	changelogCmd.Flags().StringVar(&file, "file", "", "CHANGELOG.md")
	changelogCmd.Flags().StringVar(&org, "org", "knabben", "Github owner or org")
	changelogCmd.Flags().StringVar(&repo, "repo", "", "Github repo")
	changelogCmd.Flags().StringVar(&token, "token", "./token", "Github token file (optional)")

	changelogCmd.Flags().StringVar(&label, "label", "release-note", "Release-note label")

	changelogCmd.MarkFlagRequired("file")
	changelogCmd.MarkFlagRequired("org")
	changelogCmd.MarkFlagRequired("repo")
}
