package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var (
	token string
	org   string
	repo  string
	file  string

	rootCmd = &cobra.Command{
		Use:   "chrn",
		Short: "A Changelog generator",
		Long:  ``,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
