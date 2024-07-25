/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

// toJSON converts the given object to a JSON string representation.
// If an error occurs during marshaling, an empty string is returned.
// This is taken from https://github.com/intel/tfortools
func toJSON(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}

// prefix can be used to create a list separated by s and the very first
// element is not prefixed.
func prefix(s string) func() string {
	i := -1
	return func() string {
		i++
		if i == 0 {
			return ""
		}
		return s
	}
}

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

	if result, err = o3r.Get(args); err != nil {
		return err
	}

	if cmd.Flags().Changed("format") {
		if helper.prettyPrint() {
			return errors.New("you can't use --pretty and --format at the same time")
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return fmt.Errorf("unable to get the format string from the command line: %w", err)
		}
		var inputData interface{}
		if err = json.Unmarshal([]byte(result.String()), &inputData); err != nil {
			return fmt.Errorf("unable to unmarshal the JSON data from the 'get' call: %w", err)
		}
		templateFunctions := template.FuncMap{
			"toJSON": toJSON,
			"prefix": prefix,
		}
		tmpl, err := template.New("output").Funcs(templateFunctions).Parse(format)
		if err != nil {
			return fmt.Errorf("unable to parse the template: %w", err)
		}

		if err := tmpl.Execute(os.Stdout, inputData); err != nil {
			return fmt.Errorf("unable to execute the template: %w", err)
		}
		return nil
	}

	if err := helper.printJSONResult(result.String()); err != nil {
		return err
	}
	return nil
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <pointers>",
	Short: "Retrieve the JSON configuration from the device",
	Long: `The OVP8xx get call accepts a list of JSON pointer like queries.
Valid queries are for example:

- The empty string "" which queries the complete configuration
- To query all ports including all sub elements the query "/ports" can be used.

In contrast to the concept of a JSON pointer the OVP8xx does not response with the data
the pointer is pointing to, it returns the full object hierarchy with the encapsulating
object paths.

A query of the name of the "port6" (/ports/port6/info/name) not just returns the object of that port,
it also keeps the hierarchy intact:

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
	getCmd.Flags().String("format", "", "Specify an alternative format for the JSON output")
}
