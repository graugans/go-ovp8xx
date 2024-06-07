package swupdater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type SWUpdater struct {
	hostName  string
	port      int
	path      string
	urlUpload string
	urlStatus string
	done      chan error
}

func NewSWUpdater(hostName, path string, port int) *SWUpdater {
	return &SWUpdater{
		hostName:  hostName,
		port:      port,
		path:      path,
		urlUpload: fmt.Sprintf("http://%s:%d%s/upload", hostName, port, path),
		urlStatus: fmt.Sprintf("ws://%s:%d%s/ws", hostName, port, path),
		done:      make(chan error),
	}
}

func (s *SWUpdater) upload(image io.Reader, timeout time.Duration) error {
	req, err := http.NewRequest("POST", s.urlUpload, image)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot upload software image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot upload software image: status code %d", resp.StatusCode)
	}

	return nil
}

func (s *SWUpdater) waitForFinished() {
	c, _, err := websocket.DefaultDialer.Dial(s.urlStatus, nil)
	if err != nil {
		s.done <- fmt.Errorf("cannot connect to websocket: %w", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			s.done <- fmt.Errorf("cannot read message from websocket: %w", err)
			return
		}

		data := make(map[string]string)
		err = json.Unmarshal(message, &data)
		if err != nil {
			continue
		}

		if data["type"] != "message" {
			continue
		}

		if data["text"] == "SWUPDATE successful" {
			s.done <- nil
			return
		}
		if data["text"] == "Installation failed" {
			s.done <- errors.New("installation failed")
			return
		}
	}
}

func (s *SWUpdater) Update(image io.Reader, timeout time.Duration) error {
	go s.waitForFinished()
	go func() {
		err := s.upload(image, timeout)
		if err != nil {
			s.done <- err
		}
	}()

	select {
	case err := <-s.done:
		if err != nil {
			return err
		}
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}
