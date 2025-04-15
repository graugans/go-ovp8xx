/*
Copyright © 2024 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"text/template"
	"time"

	"os"
	"strings"

	"github.com/graugans/go-ovp8xx/v2/pkg/pcic"
	"github.com/spf13/cobra"
)

// PCICReceiver represents a receiver for PCIC data.
type PCICReceiver struct {
	frame           pcic.Frame               // The PCIC frame.
	notificationMsg pcic.NotificationMessage // The notification message.
	errorMsg        pcic.ErrorMessage        // The error message.
	framecount      int64                    // The count of frames received.
	dump            bool                     // A variable to store the --dump flag
	tmpl            *template.Template       // Parsed template for dumping chunks
}

var chunkTemplate = `Chunk {{.Index}}:
  Type: {{.Type}}
  Size: {{.Size}} bytes
  Frame Count: {{.FrameCount}}
  Status: {{.Status}}
  Timestamp: {{.Timestamp}}
  Data Hexdump:
{{hexdump .Bytes}}
`

var templateString string // Global variable to store the --template flag value

// Result is a method of the PCICReceiver struct that sets the received frame and increments the framecount.
// It takes a pcic.Frame as a parameter.
func (r *PCICReceiver) Result(frame pcic.Frame) {
	r.frame = frame
	if r.dump { // Only dump if the --dump flag is set
		// Iterate over the chunks in the frame
		for i, chunk := range frame.Chunks {
			data := struct {
				Index      int
				Type       uint32
				Size       int
				FrameCount uint32
				Status     uint32
				Timestamp  string
				Bytes      []byte
			}{
				Index:      i,
				Type:       uint32(chunk.Type()),
				Size:       chunk.Size(),
				FrameCount: chunk.FrameCount(),
				Status:     chunk.Status(),
				Timestamp:  chunk.TimeStamp().String(),
				Bytes:      chunk.Bytes(),
			}

			err := r.tmpl.Execute(os.Stdout, data)
			if err != nil {
				fmt.Printf("Error executing template: %v\n", err)
			}
		}
	} else {
		fmt.Printf("Frame count: %d\n", r.framecount)
	}
	r.framecount++
}

// Error handles the error message received from the PCIC.
// It sets the errorMsg field of the PCICReceiver struct and prints the error message.
func (r *PCICReceiver) Error(msg pcic.ErrorMessage) {
	r.errorMsg = msg
	fmt.Printf("Error: <%d>: %s\n", msg.ID, msg.Message)
}

// Notification is a method of the PCICReceiver type that handles incoming notification messages.
// It updates the notificationMsg field of the receiver and prints the message to the console.
func (r *PCICReceiver) Notification(msg pcic.NotificationMessage) {
	r.notificationMsg = msg
	fmt.Printf("Notification: %v\n", msg)
}

// pcicCommand is a function that handles the execution of the "pcic" command.
// It initializes a PCICReceiver, creates a helper, and establishes a connection to the PCIC client.
// It then continuously processes incoming data using the PCIC client and the testHandler.
// If an error occurs during any of these steps, it is returned.
// Returns nil if the function completes successfully.
func pcicCommand(cmd *cobra.Command, args []string) error {
	var testHandler = &PCICReceiver{}
	var err error

	// Retrieve the slice of commands
	cmds, err := cmd.Flags().GetStringSlice("cmd")
	if err != nil {
		return err
	}

	// Check if the dump flag is set
	if testHandler.dump, err = cmd.Flags().GetBool("dump"); err != nil {
		return err
	}

	// Parse the template and store it in the receiver
	if testHandler.dump {
		funcMap := template.FuncMap{
			"hexdump": func(data []byte) string {
				return hex.Dump(data) // Dump all bytes
			},
			"hexdump_range": func(data []byte, start int, length int) string {
				if start < 0 || start >= len(data) {
					return "Invalid start index"
				}
				if length == -1 || start+length > len(data) {
					length = len(data) - start // Adjust length to dump until the end
				}
				return hex.Dump(data[start : start+length])
			},
		}

		// Use the provided template string or the default one
		if templateString == "" {
			templateString = chunkTemplate
		} else {
			// Replace literal `\n` with actual newlines
			templateString = strings.ReplaceAll(templateString, `\n`, "\n")
		}

		testHandler.tmpl, err = template.New("chunk").Funcs(funcMap).Parse(templateString)
		if err != nil {
			return fmt.Errorf("error parsing template: %v", err)
		}
	}

	helper, err := NewHelper(cmd)
	if err != nil {
		return err
	}

	pcic, err := pcic.NewPCICClient(
		pcic.WithTCPClient(helper.hostname(), helper.remotePort()),
	)
	var wg sync.WaitGroup
	wg.Add(1) // We're going to wait for one goroutine

	go func() {
		defer wg.Done() // This will be called when the goroutine finishes
		for {
			err = pcic.ProcessIncomming(testHandler)
			if err != nil {
				// An error occurred, we break the loop
				break
			}
		}
	}()

	// Execute the commands
	for _, cmd := range cmds {
		prefix := fmt.Sprintf(" %s # ", cmd)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		response, err := pcic.Send(ctx, []byte(cmd))
		if err != nil {
			cancel()
			return fmt.Errorf("failed to send command: %v", err)
		}
		if len(response) >= 9 { // Ensure there are at least 9 bytes
			lengthStr := string(response[:9])      // Convert the first 9 bytes to a string
			length, err := strconv.Atoi(lengthStr) // Convert the string to an integer
			if err != nil {
				// Response does not start with the length, print the whole response
				fmt.Println(prefix, string(response))
			} else {
				if len(response) >= 9+length {
					// Strip the first 9 bytes and print the rest up to the specified length
					fmt.Println(string(response[9 : 9+length]))
				} else {
					cancel()
					return fmt.Errorf("response too short: %s", string(response))
				}
			}
		} else {
			fmt.Println(prefix, string(response))
		}
		cancel()
	}
	// Wait for the goroutine to be finished
	wg.Wait()
	return err
}

// pcicCmd represents the pcic command
var pcicCmd = &cobra.Command{
	Use:   "pcic",
	Short: "Create a PCIC connection to the device",
	RunE:  pcicCommand,
}

func init() {
	rootCmd.AddCommand(pcicCmd)
	pcicCmd.Flags().Uint16("port", 50010, "The port to connect to")
	pcicCmd.Flags().StringSlice("cmd", []string{}, "Commands to be send to the device, can be specified multiple times. All commands will be executed in order")
	pcicCmd.Flags().Bool("dump", false, "Dump frame and chunk data")
	pcicCmd.Flags().StringVar(&templateString, "template", "", "Custom template for dumping frame and chunk data")
}
