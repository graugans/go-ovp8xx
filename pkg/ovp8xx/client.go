package ovp8xx

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

type (
	ClientOption func(c *Client)
	Client       struct {
		host string
		url  string
	}
	DiagnosisClient struct {
		host string
		url  string
	}
)

// NewClient creates a new OVP8xx client with the provided options.
// The opts parameter is a variadic parameter that allows specifying
// multiple client options.
// Example usage:
//
//	client := NewClient(WithHost("192.168.47.11"))
//	// ...
func NewClient(opts ...ClientOption) *Client {
	// Initialise with default values
	client := &Client{
		host: GetEnv("OVP8XX_IP", "192.168.0.69"),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}
	client.url = fmt.Sprintf("http://%s/api/rpc/v1/com.ifm.efector/", client.host)
	return client
}

// WithHost sets the host for the OVP8xx client.
// It returns a ClientOption function that can be used to configure the client.
func WithHost(host string) ClientOption {
	return func(c *Client) {
		c.host = host
	}
}

// GetDiagnosticClient returns a new instance of DiagnosisClient that can be used to perform diagnostic operations on the OVP8xx device.
func (device *Client) GetDiagnosticClient() *DiagnosisClient {
	client := &DiagnosisClient{}
	client.host = device.host
	client.url = fmt.Sprintf("http://%s/api/rpc/v1/com.ifm.diagnostic/", client.host)
	return client
}

// WaitForConfig represents the configuration for waiting for a specific stage.
type WaitForConfig struct {
	stage string // The stage to wait for.
}

// WaitForOption is a function type that is used as an option for configuring the behavior of the WaitFor function.
// It takes a pointer to a WaitForConfig struct as a parameter and can be used to modify its properties.
type WaitForOption func(c *WaitForConfig)

// AndPortsAreOnline returns a WaitForOption function that sets the stage to "ports".
func AndPortsAreOnline() WaitForOption {
	return func(w *WaitForConfig) {
		w.stage = "ports"
	}
}

// AndAppsAreOnline returns a WaitForOption function that sets the stage to "applications".
func AndAppsAreOnline() WaitForOption {
	return func(w *WaitForConfig) {
		w.stage = "applications"
	}
}

// IsAvailable checks if the client is available by making a XML-RPC get request to query the "/device" object.
// This is useful to wait until a device is ready for communication.
// It returns a boolean indicating the availability status and an error if any.
// The function uses a timeout duration to limit the execution time of the request.
// Example usage:
//
//	ok, err := device.IsAvailable(time.Duration(timeout) * time.Second, ovp8xx.AndPortsAreOnline())
//	// ...
func (d *Client) IsAvailable(timeout time.Duration, opts ...WaitForOption) (bool, error) {
	var err error
	proc := make(chan struct{}, 1)
	conf := *NewConfig()

	waitingFor := &WaitForConfig{
		stage: "device",
	}

	// Apply wait options
	for _, opt := range opts {
		opt(waitingFor)
	}

	// result is a struct that represents the response from the OVP8xx device.
	// It contains information about the device's diagnostic data, such as the configuration initialization stages.
	result := struct {
		Device struct {
			Diagnostic struct {
				ConfInitStages []string `json:"confInitStages"`
			} `json:"diagnostic"`
		} `json:"device"`
	}{}

	go func() {
		for {
			if conf, err = d.Get([]string{"/device/diagnostic/confInitStages"}); err != nil {
				// In case of an error retry, regardless of an timeout
				continue
			}
			// Unmarshal the data into the result struct
			if err = json.Unmarshal([]byte(conf.String()), &result); err != nil {
				// In case of an error retry until the timeout
				continue
			}
			// Check if the device is ready
			if slices.Contains(result.Device.Diagnostic.ConfInitStages, waitingFor.stage) {
				// we are done, the get call was successful
				proc <- struct{}{}
			}
		}
	}()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return false, fmt.Errorf("timeout occurred while checking if the device is available: %w", ctx.Err())
	case <-proc:
		return true, err
	}
}
