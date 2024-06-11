package swupdater

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type SWUpdater struct {
	hostName  string
	port      uint16
	urlUpload string
	urlStatus string
}

func NewSWUpdater(hostName string, port uint16) *SWUpdater {
	return &SWUpdater{
		hostName:  hostName,
		port:      port,
		urlUpload: fmt.Sprintf("http://%s:%d/upload", hostName, port),
		urlStatus: fmt.Sprintf("ws://%s:%d/ws", hostName, port),
	}
}
func (s *SWUpdater) upload(filename string, timeout time.Duration) error {
	image, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	fmt.Printf("Uploading software image to %s\n", s.urlUpload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("cannot create form file: %w", err)
	}

	_, err = io.Copy(part, image)
	if err != nil {
		return fmt.Errorf("cannot write to form file: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("cannot close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", s.urlUpload, bytes.NewReader(body.Bytes()))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))

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

		if data["text"] == "SWUPDATE successful" {
			done <- nil
			return
		}
		if data["text"] == "Installation failed" {
			done <- errors.New("installation failed")
			return
		}
	}
}

func (s *SWUpdater) Update(filename string, timeout time.Duration) error {
	done := make(chan error)
	go s.waitForFinished(done)
	err := s.upload(filename, timeout)
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
