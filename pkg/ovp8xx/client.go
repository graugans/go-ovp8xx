package ovp8xx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"net/url"
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

// DefaultHost is the factory default IP address of the device.
const DefaultHost = "192.168.0.69"

// NewClient creates a new OVP8xx client with the provided options.
// The opts parameter is a variadic parameter that allows specifying
// multiple client options. Without WithHost the client connects to
// DefaultHost. The environment is not consulted; use the variant package to
// resolve the host from environment variables.
// Example usage:
//
//	client := NewClient(WithHost("192.168.47.11"))
//	// ...
func NewClient(opts ...ClientOption) *Client {
	// Initialise with default values
	client := &Client{
		host: DefaultHost,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}
	client.url = rpcURL(client.host, "/api/rpc/v1/com.ifm.efector/")
	return client
}

// rpcURL builds the URL of an XML-RPC endpoint. The host may be a hostname,
// an IPv4 or IPv6 address (optionally with a zone, e.g. "fe80::1%eth0") or any
// of these combined with a port, e.g. "[::1]:8080".
func rpcURL(host, path string) string {
	// A bare IPv6 address needs to be enclosed in brackets
	if addr, err := netip.ParseAddr(host); err == nil && addr.Is6() {
		host = "[" + host + "]"
	}
	u := url.URL{Scheme: "http", Host: host, Path: path}
	return u.String()
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
	client.url = rpcURL(client.host, "/api/rpc/v1/com.ifm.diagnostic/")
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

// availabilityPollInterval is the pause between two availability checks.
const availabilityPollInterval = 250 * time.Millisecond

// IsAvailable checks if the client is available by making a XML-RPC get request to query the "/device" object.
// This is useful to wait until a device is ready for communication.
// It returns a boolean indicating the availability status and an error if any.
// The function uses a timeout duration to limit the execution time of the request.
// Example usage:
//
//	ok, err := device.IsAvailable(time.Duration(timeout) * time.Second, ovp8xx.AndPortsAreOnline())
//	// ...
func (d *Client) IsAvailable(timeout time.Duration, opts ...WaitForOption) (bool, error) {
	waitingFor := &WaitForConfig{
		stage: "device",
	}

	// Apply wait options
	for _, opt := range opts {
		opt(waitingFor)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// The check runs in its own goroutine, because a single request may block
	// longer than the timeout. The goroutine stops once the context is done.
	available := make(chan struct{}, 1)
	go func() {
		for {
			if d.reachedStage(waitingFor.stage) {
				available <- struct{}{}
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(availabilityPollInterval):
			}
		}
	}()

	select {
	case <-ctx.Done():
		return false, fmt.Errorf("timeout occurred while checking if the device is available: %w", ctx.Err())
	case <-available:
		return true, nil
	}
}

// reachedStage returns true if the device reports the given configuration
// initialization stage.
func (d *Client) reachedStage(stage string) bool {
	conf, err := d.Get([]string{"/device/diagnostic/confInitStages"})
	if err != nil {
		return false
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
	if err := json.Unmarshal([]byte(conf.String()), &result); err != nil {
		return false
	}
	return slices.Contains(result.Device.Diagnostic.ConfInitStages, stage)
}
