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
	hostName      string                     // The hostname of the updater.
	port          uint16                     // The port number of the updater.
	urlUpload     string                     // The URL for uploading software updates.
	urlStatus     string                     // The URL for checking the status of software updates.
	notifications chan SWUpdaterNotification // A channel for receiving notifications.
	ws            *websocket.Conn
}

type SWUpdaterNotification map[string]string

// NewSWUpdater creates a new instance of SWUpdater with the specified host name and port.
func NewSWUpdater(hostName string, port uint16, notifications chan SWUpdaterNotification) *SWUpdater {
	return &SWUpdater{
		hostName:      hostName,
		port:          port,
		urlUpload:     fmt.Sprintf("http://%s:%d/upload", hostName, port),
		urlStatus:     fmt.Sprintf("ws://%s:%d/ws", hostName, port),
		notifications: notifications,
	}
}

// Upload performs the upload of the specified file.
// The filename parameter specifies the name of the file to be uploaded.
// Returns an error if the upload fails.
func (s *SWUpdater) upload(filename string) error {
	s.statusUpdate(fmt.Sprintf("Uploading software image to %s\n", s.urlUpload))
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
	err = ms.WriteReader(fieldname, filename, fileInfo.Size(), file)
	if err != nil {
		return fmt.Errorf("cannot write reader: %w", err)
	}

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

	for {
		_, message, err := s.ws.ReadMessage()
		if err != nil {
			done <- fmt.Errorf("cannot read message from websocket: %w", err)
			return
		}

		data := make(SWUpdaterNotification)
		err = json.Unmarshal(message, &data)
		if err != nil {
			done <- fmt.Errorf("cannot unmarshal message: %w", err)
			return
		}
		// Send notification to channel
		if s.notifications != nil {
			s.notifications <- data
		}
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

func (s *SWUpdater) connect() error {
	var err error
	s.ws, _, err = websocket.DefaultDialer.Dial(s.urlStatus, nil)
	if err != nil {
		return fmt.Errorf("unable to connect to the status socket: %w", err)
	}
	return err
}

func (s *SWUpdater) disconnect() {
	s.ws.Close()
}

// statusUpdate updates the status of the SWUpdater.
// It sends a notification to the channel with the provided status.
func (s *SWUpdater) statusUpdate(status string) {
	notification := make(SWUpdaterNotification)
	notification["swupdater"] = status
	// Send notification to channel
	if s.notifications != nil {
		s.notifications <- notification
	}
}

// Update uploads a software image and waits for the update process to finish.
// It takes a filename string and a timeout duration as parameters.
// It returns an error if the upload fails, or if the operation times out.
func (s *SWUpdater) Update(filename string, connectionTimeout, timeout time.Duration) error {
	done := make(chan error)
	start := time.Now()
	s.statusUpdate("Waiting for the Device to become ready...")
	// Retry connection until successful or connectionTimeout occurs
	for {
		err := s.connect()
		if err == nil {
			s.statusUpdate("Device is ready now")
			break
		}
		if time.Since(start) > connectionTimeout {
			return fmt.Errorf("connection timeout: %w", err)
		}
		time.Sleep(3 * time.Second) // wait for a second before retrying
	}
	defer s.disconnect()

	s.statusUpdate("Starting the Software Update process...")
	go s.waitForFinished(done)
	err := s.upload(filename)
	if err != nil {
		return fmt.Errorf("cannot upload software image: %w", err)
	}

	select {
	case err := <-done:
		close(s.notifications) // Close the channel to signal the end of notifications
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		return nil
	case <-time.After(timeout):
		return errors.New("a timeout occurred while waiting for the update to finish")
	}
}
