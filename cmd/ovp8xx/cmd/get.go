/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func getCommand(cmd *cobra.Command, args []string) error {
	var result ovp8xx.Config
	var err error
	helper, err := NewHelper(cmd)
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(helper.hostname()),
	)

	if result, err = o3r.Get(helper.jsonPointers()); err != nil {
		return err
	}
	if err := helper.printJSONResult(result.String()); err != nil {
		return err
	}
	return nil
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve the JSON configuration from the device",
	Long: `The OVP8xx get call accepts a list of JSON pointer like queries.
Valid queries are for example:

- The empty string "" which queries the complete configuration
- To query all ports including all sub elements the query "/ports" can be used.

In contrast to the concept of a JSON pointer the OVP8xx does not response with the data
the pointer is pointing to, it returns the full object hirachie with the encapsulating
object paths.

A query of the name of the "port6" (/ports/port6/info/name) not just returns the object of that port,
it also keeps the hirachy intact:

{
	"ports":
		"port6": {
			"info": {
				"name":"Front Left"
			}
		}
}

This allows one to use the response of a "get" command to directly feed it into a "set" command

NOTE: This command only modifies temporary data, any changes will be lost after a reboot or power off.
`,
	RunE: getCommand,
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringSliceP("pointer", "p", []string{""}, "A JSON pointer to be queried")
	getCmd.Flags().Bool("pretty", false, "Pretty print the JSON received from the device")
}
