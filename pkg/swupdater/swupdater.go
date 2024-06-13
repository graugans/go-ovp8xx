package swupdater

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/technoweenie/multipartstreamer"
)

// SWUpdater represents a software updater.
type SWUpdater struct {
	hostName  string // The hostname of the updater.
	port      uint16 // The port number of the updater.
	urlUpload string // The URL for uploading software updates.
	urlStatus string // The URL for checking the status of software updates.
}

// NewSWUpdater creates a new instance of SWUpdater with the specified host name and port.
func NewSWUpdater(hostName string, port uint16) *SWUpdater {
	return &SWUpdater{
		hostName:  hostName,
		port:      port,
		urlUpload: fmt.Sprintf("http://%s:%d/upload", hostName, port),
		urlStatus: fmt.Sprintf("ws://%s:%d/ws", hostName, port),
	}
}

// Upload performs the upload of the specified file.
// The filename parameter specifies the name of the file to be uploaded.
// Returns an error if the upload fails.
func (s *SWUpdater) upload(filename string) error {
	fmt.Printf("Uploading software image to %s\n", s.urlUpload)
	const fieldname string = "file"

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot get file info: %w", err)
	}

	ms := multipartstreamer.New()
	ms.WriteReader(fieldname, filename, fileInfo.Size(), file)

	req, _ := http.NewRequest("POST", s.urlUpload, nil)
	ms.SetupRequest(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}
	defer resp.Body.Close()
	return err
}

// waitForFinished waits for the SWUpdater process to finish by listening to a WebSocket connection.
// It continuously reads messages from the WebSocket and checks for specific conditions to determine
// if the SWUpdater process has completed successfully or has failed.
//
// Parameters:
//   - done: A channel used to signal the completion of the SWUpdater process. If the process finishes
//     successfully, nil is sent to the channel. If the process fails, an error is sent to the channel.
//
// Returns:
//
//	None
//
// Example usage:
//
//	done := make(chan error)
//	go s.waitForFinished(done)
//	err := <-done
//	if err != nil {
//	  // Handle error
//	} else {
//	  // SWUpdater process completed successfully
//	}
func (s *SWUpdater) waitForFinished(done chan error) {
	c, _, err := websocket.DefaultDialer.Dial(s.urlStatus, nil)
	if err != nil {
		done <- fmt.Errorf("cannot connect to websocket: %w", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			done <- fmt.Errorf("cannot read message from websocket: %w", err)
			return
		}

		data := make(map[string]string)
		err = json.Unmarshal(message, &data)
		if err != nil {
			done <- fmt.Errorf("cannot unmarshal message: %w", err)
			return
		}
		fmt.Println("Raw JSON: ", data)
		if data["type"] != "message" {
			continue
		}

		if strings.Contains(data["text"], "SWUPDATE successful") {
			done <- nil
			return
		}
		if strings.Contains(data["text"], "Installation failed") {
			done <- errors.New("installation failed")
			return
		}
	}
}

// Update uploads a software image and waits for the update process to finish.
// It takes a filename string and a timeout duration as parameters.
// It returns an error if the upload fails, or if the operation times out.
func (s *SWUpdater) Update(filename string, timeout time.Duration) error {
	done := make(chan error)
	go s.waitForFinished(done)
	err := s.upload(filename)
	if err != nil {
		return fmt.Errorf("cannot upload software image: %w", err)
	}

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}
