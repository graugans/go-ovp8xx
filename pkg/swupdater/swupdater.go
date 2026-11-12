package swupdater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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
	urlDeviceInfo string                     // The URL used to check whether the service is reachable.
	notifications chan SWUpdaterNotification // A channel for receiving notifications.
	ws            *websocket.Conn
}

type SWUpdaterNotification map[string]string

// NewSWUpdater creates a new instance of SWUpdater with the specified host name and port.
func NewSWUpdater(hostName string, port uint16, notifications chan SWUpdaterNotification) *SWUpdater {
	return &SWUpdater{
		hostName:      hostName,
		port:          port,
		urlUpload:     serviceURL("http", hostName, port, "/upload"),
		urlStatus:     serviceURL("ws", hostName, port, "/ws"),
		urlDeviceInfo: serviceURL("http", hostName, port, "/device-info"),
		notifications: notifications,
	}
}

// serviceURL builds the URL of a SWUpdate endpoint. IPv6 addresses are
// enclosed in brackets, an optional zone (e.g. "fe80::1%eth0") is escaped.
func serviceURL(scheme, host string, port uint16, path string) string {
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	u := url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   path,
	}
	return u.String()
}

// requiredAvailabilityChecks is the number of consecutive successful checks
// required before the SWUpdate service is considered available. A single
// answer is not sufficient, the service may still be restarted while booting.
const requiredAvailabilityChecks = 3

// availabilityRequestTimeout limits a single availability check, so a
// device which accepts the connection but does not answer is polled again.
const availabilityRequestTimeout = 3 * time.Second

// availabilityPollInterval is the pause between two availability checks.
var availabilityPollInterval = time.Second

// IsAvailable waits until the SWUpdate service of the device is reachable,
// i.e. the device is running in the SWUpdate (recovery) mode. The service is
// considered available after it answered three consecutive requests. The
// device is polled until it is available or ctx is done.
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
//	defer cancel()
//	ok, err := swu.IsAvailable(ctx)
//	// ...
func (s *SWUpdater) IsAvailable(ctx context.Context) (bool, error) {
	ticker := time.NewTicker(availabilityPollInterval)
	defer ticker.Stop()
	stable := 0
	for {
		if s.deviceInfoReachable(ctx) {
			stable++
			if stable >= requiredAvailabilityChecks {
				return true, nil
			}
		} else {
			stable = 0
		}
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("SWUpdate service is not available: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// deviceInfoReachable returns true if the SWUpdate service answers a request
// for the device information successfully.
func (s *SWUpdater) deviceInfoReachable(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, availabilityRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.urlDeviceInfo, nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return false
	}
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// Upload performs the upload of the specified file.
// The filename parameter specifies the name of the file to be uploaded.
// Returns an error if the upload fails.
func (s *SWUpdater) upload(filename string) (err error) {
	basename := filepath.Base(filename)
	s.statusUpdate(
		fmt.Sprintf("Uploading software image: %s to %s\n", basename, s.urlUpload),
	)
	const fieldname string = "file"

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot get file info: %w", err)
	}

	ms := multipartstreamer.New()
	err = ms.WriteReader(fieldname, filename, fileInfo.Size(), file)
	if err != nil {
		return fmt.Errorf("cannot write reader: %w", err)
	}

	req, err := http.NewRequest("POST", s.urlUpload, nil)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	ms.SetupRequest(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("the upload was rejected with status: %s", resp.Status)
	}
	return nil
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

func (s *SWUpdater) disconnect() error {
	return s.ws.Close()
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
func (s *SWUpdater) Update(filename string, connectionTimeout, timeout time.Duration) (err error) {
	// The channel is buffered, so waitForFinished can terminate even if
	// Update already returned, e.g. after a timeout.
	done := make(chan error, 1)
	// Retry connection until successful or connectionTimeout occurs
	online, err := s.waitForOnline(connectionTimeout)
	if !online {
		return err
	}
	// close the websocket after the Update operation
	defer func() {
		if closeErr := s.disconnect(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	s.statusUpdate("Starting the Software Update process...")
	go s.waitForFinished(done)
	err = s.upload(filename)
	if err != nil {
		return fmt.Errorf("cannot upload software image: %w", err)
	}

	select {
	case err := <-done:
		// waitForFinished returned, so nobody sends notifications anymore.
		// Close the channel to signal the end of notifications.
		if s.notifications != nil {
			close(s.notifications)
		}
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		return nil
	case <-time.After(timeout):
		return errors.New("a timeout occurred while waiting for the update to finish")
	}
}

func (s *SWUpdater) waitForOnline(connectionTimeout time.Duration) (bool, error) {
	start := time.Now()
	s.statusUpdate("Waiting for the Device to become ready...")

	for {
		err := s.connect()
		if err == nil {
			s.statusUpdate("Device is ready now")
			break
		}
		if time.Since(start) > connectionTimeout {
			return false, fmt.Errorf("connection timeout: %w", err)
		}
		// Retry after 3 seconds
		time.Sleep(3 * time.Second)
	}
	return true, nil
}

// Restart reboots the device by sending a POST request to the restart endpoint.
func (s *SWUpdater) Restart(timeout time.Duration) error {
	// Construct the URL for the restart endpoint
	restartURL := serviceURL("http", s.hostName, s.port, "/restart")

	online, err := s.waitForOnline(timeout)
	if !online {
		return err
	}
	// close the websocket after the Restart operation
	defer func() {
		if closeErr := s.disconnect(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Create a POST request with an empty body
	req, err := http.NewRequest("POST", restartURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request (%s): %w", restartURL, err)
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send restart request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("the SWUpdate service is not available at the moment, please try again later")
	}

	// Check if the response status code indicates success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("restart request (%s) failed with status code: %d", restartURL, resp.StatusCode)
	}

	return nil
}
