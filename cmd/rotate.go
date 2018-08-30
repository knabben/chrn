package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	ssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	bump string

	rotateCmd = &cobra.Command{
		Use:   "rotate",
		Short: "Rotate changelog to unreleased flag",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			gh := NewClientBasedOnToken(org, token)
			dir, fileName := filepath.Split(file)

			auth, err := ssh.NewPublicKeysFromFile(
				"git",
				fmt.Sprintf("%v/.ssh/id_rsa", os.Getenv("HOME")), "",
			)
			CheckIfError(err)

			latestTag, err := BumpLastTagVersion(gh)
			if err != nil {
				latestTag = "0.0.1"
			}

			RunRotate(latestTag, dir, fileName, gh, auth)
		},
	}
)

// RunRotate rotate changelog with a new PR
func RunRotate(latestTag string, dir string, fileName string, gh *GithubClient, auth *ssh.PublicKeys) {
	color.Cyan(fmt.Sprintf(">> Using %v as the latest version for this project.", latestTag))

	color.Cyan(fmt.Sprintf(">> Creating a new branch.. release/%s", latestTag))
	CreateBranchAndCheckout(latestTag, dir, auth)
	color.Green(fmt.Sprintf(">> Branch created."))

	color.Cyan(fmt.Sprintf(">> Bumping changelog file"))
	UpdateFiles(latestTag)
	color.Green(fmt.Sprintf(">> Changelog/version modified."))

	color.Cyan(fmt.Sprintf(">> Commiting to the branch"))
	AddCommitBranch(dir, fileName, latestTag)
	color.Green(fmt.Sprintf(">> Branch commited."))

	for {
		color.Yellow("Opening PR, do you Want to proceed? [y|yes]")

		reader := bufio.NewReader(os.Stdin)
		continueProcess, _ := reader.ReadString('\n')

		if continueProcess == "y\n" || continueProcess == "yes\n" {
			PushOpenPR(dir, gh, latestTag, auth)
			os.Exit(0)

		} else if continueProcess == "n\n" || continueProcess == "no\n" {
			color.Red(fmt.Sprintf(">> Cya."))
			os.Exit(1)
		}
	}

}

// PushOpenPR sends branch to remote origin and create a new PR
func PushOpenPR(repoPath string, gh *GithubClient, tag string, auth *ssh.PublicKeys) {
	repository, _ := git.PlainOpen(repoPath)

	color.Cyan(fmt.Sprintf(">> Pushing to remote..."))
	err := repository.Push(&git.PushOptions{
		Auth:       auth,
		RemoteName: "origin",
	})
	CheckIfError(err)

	color.Cyan(fmt.Sprintf(">> Creating a new PR..."))
	pr, err := gh.CreateNewPR(repo, tag, org)
	CheckIfError(err)

	color.Green(fmt.Sprintf(pr.GetHTMLURL()))
	color.Green(fmt.Sprintf(">> PR openend."))
}

// CommitPushBranch add modified file commit it and push branch
func AddCommitBranch(repoPath string, fileName string, tag string) {
	repository, _ := git.PlainOpen(repoPath)
	worktree, _ := repository.Worktree()

	_, err := worktree.Add(fileName)
	CheckIfError(err)

	_, err = worktree.Add(versionFile)
	CheckIfError(err)

	_, err = worktree.Commit(fmt.Sprintf("Release %v", tag), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Amim Knabben",
			Email: "amim.knabben@gmail.com",
			When:  time.Now(),
		},
	})
	CheckIfError(err)
}

// CreateBranchAndCheckout creates a new branch checkout
func CreateBranchAndCheckout(tag string, repoPath string, auth *ssh.PublicKeys) {
	repository, _ := git.PlainOpen(repoPath)
	worktree, _ := repository.Worktree()

	color.Cyan(fmt.Sprintf("Pulling repo..."))
	err := worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err != nil {
		color.Red(fmt.Sprintf("%v", err))
	}

	headRef, err := repository.Head()
	CheckIfError(err)

	branch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/release/%v", tag))
	ref := plumbing.NewHashReference(branch, headRef.Hash())
	err = repository.Storer.SetReference(ref)
	CheckIfError(err)

	err = worktree.Checkout(&git.CheckoutOptions{Branch: branch})
	CheckIfError(err)
}

// UpdateFiles update files with new content:
// allowed files are changelog and version.py
func UpdateFiles(tag string) {
	// CHANGELOG.md
	chFile, err := os.Open(file)
	CheckIfError(err)
	defer chFile.Close()
	changeLogNewLines := BumpChangelogFile(chFile, tag)
	WriteLinesNewFile(changeLogNewLines, file)

	// version.py
	WriteLinesNewFile(
		[]string{fmt.Sprintf("VERSION = \"%s\"", tag)}, versionFile,
	)
}

// BumpLastTagVersion finds the last release tag and bump it
func BumpLastTagVersion(gh *GithubClient) (string, error) {
	latest, err := gh.GetLatestRelease(repo)
	if err != nil {
		return "", err
	}

	v, err := semver.Make(strings.Replace(latest, "v", "", -1))

	switch bump {
	case "minor":
		v.Minor += 1
		v.Patch = 0

	case "major":
		v.Major += 1
		v.Minor = 0
		v.Patch = 0
	case "patch":
		v.Patch += 1
	}

	return v.String(), nil
}

// WriteLinesNewFile writes a line slice to a file
func WriteLinesNewFile(lines []string, filename string) error {
	newFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	for _, line := range lines {
		newFile.WriteString(line)
	}

	return nil
}

// BumpChangelogFile read a changelog and replace for new entries
func BumpChangelogFile(chFile *os.File, tag string) []string {
	scanner := bufio.NewScanner(chFile)
	scanner.Split(bufio.ScanLines)

	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "## [Unreleased]") {
			// Title header change.
			line = fmt.Sprintf(
				"%v\n\n## [%v] - %v",
				line,
				tag,
				time.Now().Format("2006-01-02"),
			)

		} else if strings.Contains(line, "[Unreleased]") {
			// Link change
			re, _ := regexp.Compile(`\[Unreleased\]: (?P<compare_url>.*)/(?P<old_tag>.*)...HEAD`)
			match := re.FindAllStringSubmatch(line, -1)[0]

			ghUrl := match[1]
			oldTag := strings.Replace(match[2], "v", "", -1)

			line = fmt.Sprintf(
				"[Unreleased]: %v/%v...HEAD\n[%v]: %v/%v...%v",
				ghUrl, tag, tag, ghUrl, oldTag, tag,
			)
		}

		lines = append(lines, fmt.Sprintf("%v\n", line))
	}
	return lines
}

func init() {
	rootCmd.AddCommand(rotateCmd)

	rotateCmd.Flags().StringVar(&org, "org", "knabben", "Github owner or org")
	rotateCmd.Flags().StringVar(&repo, "repo", "", "Github repo")
	rotateCmd.Flags().StringVar(&token, "token", "./token", "Github token file (optional)")
	rotateCmd.Flags().StringVar(&file, "file", "", "CHANGELOG.md")
	rotateCmd.Flags().StringVar(&bump, "bump", "minor", "Bump type [major, minor, patch]")
	rotateCmd.Flags().StringVar(&versionFile, "version", "version", "head/version.py", "Default version file")

	rotateCmd.MarkFlagRequired("file")
	rotateCmd.MarkFlagRequired("org")
	rotateCmd.MarkFlagRequired("repo")
}
