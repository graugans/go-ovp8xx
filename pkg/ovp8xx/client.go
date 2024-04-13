package ovp8xx

import (
	"context"
	"fmt"
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
//	client := NewClient(WithTimeout(10 * time.Second), WithRetry(3))
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

// IsAvailable checks if the client is available by making a XML-RPC get request to query the "/device" object.
// This is useful to wait until a device is ready for communication.
// It returns a boolean indicating the availability status and an error if any.
// The function uses a timeout duration to limit the execution time of the request.
func (d *Client) IsAvailable(timeout time.Duration) (bool, error) {
	var err error
	proc := make(chan struct{}, 1)

	go func() {
		for {
			if _, err := d.Get([]string{"/device"}); err != nil {
				// In case of an error retry, regardless of an timeout
				continue
			}
			// we are done, the get call was successful
			proc <- struct{}{}
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
