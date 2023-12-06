/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func getSchemaCommand(cmd *cobra.Command, args []string) error {

	pointers, err := cmd.Flags().GetStringSlice("pointer")
	if err != nil {
		return err
	}

	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	if result, err := o3r.GetSchema(pointers); err != nil {
		return err
	} else {
		fmt.Printf("%s\n", result)
	}
	return nil
}

// getCmd represents the get command
var getSchemaCmd = &cobra.Command{
	Use:   "getSchema",
	Short: "Retrieve the currently used JSON schema from the device",
	Long: `The OVP8xx getSchema command accepts a list of JSON pointers.
The JSON schema provides details about multiple aspects of a paramter. It
contains information like the type of a parameter and its defaults. It also
provides information weather a parameter is readOnly or not.

Due to the fact the schema can schema can grow quite big it is possible to
limit the scope of the query with a query string which is similar to a JSON pointer.

The pointer '/device/swVersion/diagnostics' for example provides this information

{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"additionalProperties": false,
	"properties": {
	  "device": {
		"$schema": "http://json-schema.org/draft-07/schema#",
		"additionalProperties": false,
		"properties": {
		  "swVersion": {
			"additionalProperties": false,
			"description": "version of software components",
			"properties": {
			  "diagnostics": {
				"readOnly": true,
				"type": "string"
			  }
			},
			"required": [
			  "kernel",
			  "l4t",
			  "firmware"
			],
			"type": "object"
		  }
		},
		"title": "O3R device configuration",
		"type": "object"
	  }
	},
	"title": "O3R configuration",
	"type": "object"
  }

When no query is provided the complete schema is returend.
`,
	RunE: getSchemaCommand,
}

func init() {
	rootCmd.AddCommand(getSchemaCmd)
	getSchemaCmd.Flags().StringSliceP("pointer", "p", []string{""}, "A JSON pointer to be queried. This can be used multiple times")
}
