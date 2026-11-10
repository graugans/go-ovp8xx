/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// SetVersionInfo sets the version information for the root command.
// It formats the version, commit, and date into a string and assigns it to the rootCmd.Version variable.
func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          currentVariant.name,
	Short:        "A command line application to interact with the ifm " + currentVariant.displayName + " series of devices",
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	envHint := currentVariant.ipEnvVar()
	if currentVariant != defaultVariant {
		envHint += " or " + defaultVariant.ipEnvVar()
	}
	rootCmd.PersistentFlags().String(
		"ip",
		currentVariant.defaultHost(os.Getenv),
		fmt.Sprintf(
			"The IP address or hostname of the %s. If not provided the default will be taken from the environment variable %s",
			currentVariant.displayName, envHint,
		),
	)
}
