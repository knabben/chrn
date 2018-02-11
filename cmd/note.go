package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// noteCmd represents the note command
var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		gh := NewClientBasedOnToken(org, token)

		chFile, err := os.Open(file)
		CheckIfError(err)
		defer chFile.Close()

		latest, notes := FirstReleaseChangelog(chFile)

		color.Cyan(fmt.Sprintf(">>> Creating a new release using tag %v and notes:", latest))
		color.Cyan(notes)

		color.Yellow("Do you Want to proceed?")
		reader := bufio.NewReader(os.Stdin)
		continueProcess, _ := reader.ReadString('\n')

		if continueProcess == "y\n" {
			err = gh.CreateNewRelease(repo, latest, notes)
			CheckIfError(err)
		}
	},
}

func FirstReleaseChangelog(chFile *os.File) (string, string) {
	scanner := bufio.NewScanner(chFile)
	scanner.Split(bufio.ScanLines)

	tag := ""
	notes := ""

	for scanner.Scan() {
		line := scanner.Text()

		re, _ := regexp.Compile(`## \[(?P<tag>.*)\] - `)
		match := re.FindAllStringSubmatch(line, -1)

		// Tag already set stop
		if len(match) >= 1 && tag != "" {
			break
		}
		if tag != "" {
			notes += fmt.Sprintf("%v\n", line)
		}
		// Try to find the first tag
		if len(match) >= 1 && tag == "" {
			tag = match[0][1]
		}
	}

	return tag, notes
}

func init() {
	rootCmd.AddCommand(noteCmd)

	noteCmd.Flags().StringVar(&file, "file", "", "CHANGELOG.md")
	noteCmd.Flags().StringVar(&org, "org", "knabben", "Github owner or org")
	noteCmd.Flags().StringVar(&repo, "repo", "", "Github repo")
	noteCmd.Flags().StringVar(&token, "token", "./token", "Github token file (optional)")

	noteCmd.MarkFlagRequired("file")
	noteCmd.MarkFlagRequired("org")
	noteCmd.MarkFlagRequired("repo")
}
