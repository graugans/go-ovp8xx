/*
Copyright © 2024 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/pcic"
	"github.com/spf13/cobra"
)

// PCICReceiver represents a receiver for PCIC data.
type PCICReceiver struct {
	frame           pcic.Frame               // The PCIC frame.
	notificationMsg pcic.NotificationMessage // The notification message.
	errorMsg        pcic.ErrorMessage        // The error message.
	framecount      int64                    // The count of frames received.
}

// Result is a method of the PCICReceiver struct that sets the received frame and increments the framecount.
// It takes a pcic.Frame as a parameter.
func (r *PCICReceiver) Result(frame pcic.Frame) {
	r.frame = frame
	fmt.Printf("Framecount: %d\n", r.framecount)
	r.framecount++
}

// Error handles the error message received from the PCIC.
// It sets the errorMsg field of the PCICReceiver struct and prints the error message.
func (r *PCICReceiver) Error(msg pcic.ErrorMessage) {
	r.errorMsg = msg
	fmt.Printf("Error: %v\n", msg)
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
	var testHandler *PCICReceiver = &PCICReceiver{}
	var err error
	helper, err := NewHelper(cmd)
	if err != nil {
		return err
	}

	pcic, err := pcic.NewPCICClient(
		pcic.WithTCPClient(helper.hostname(), helper.remotePort()),
	)
	if err != nil {
		return err
	}
	for {
		err = pcic.ProcessIncomming(testHandler)
		if err != nil {
			// An error occured, we break the loop
			break
		}
	}
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
}
