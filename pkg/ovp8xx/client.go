package ovp8xx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"time"
)

type (
	ClientOption func(c *Client)
	Client       struct {
		host string
		endpoint
	}
	DiagnosisClient struct {
		host string
		endpoint
	}
	// endpoint is a single XML-RPC endpoint of the device together with the
	// settings used for each request.
	endpoint struct {
		url        string
		httpClient *http.Client  // nil means http.DefaultClient
		timeout    time.Duration // 0 means no timeout besides the context
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

// WithTimeout limits the duration of each single XML-RPC request. A request is
// aborted when either the timeout expires or the context passed to the method
// is done. A timeout of 0 (the default) disables the limit.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithHTTPClient sets the HTTP client used for the XML-RPC requests, e.g. to
// configure a proxy or custom transport settings. By default
// http.DefaultClient is used.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// GetDiagnosticClient returns a new instance of DiagnosisClient that can be used to perform diagnostic operations on the OVP8xx device.
func (device *Client) GetDiagnosticClient() *DiagnosisClient {
	client := &DiagnosisClient{host: device.host, endpoint: device.endpoint}
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

// IsAvailable waits until the device reports the requested configuration
// initialization stage. Without options it waits for the "device" stage.
// This is useful to wait until a device is ready for communication, e.g. after
// a reboot. The device is polled until it is available or ctx is done.
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
//	defer cancel()
//	ok, err := device.IsAvailable(ctx, ovp8xx.AndPortsAreOnline())
//	// ...
func (d *Client) IsAvailable(ctx context.Context, opts ...WaitForOption) (bool, error) {
	waitingFor := &WaitForConfig{
		stage: "device",
	}

	// Apply wait options
	for _, opt := range opts {
		opt(waitingFor)
	}

	ticker := time.NewTicker(availabilityPollInterval)
	defer ticker.Stop()
	for {
		// A request blocking longer than the context allows is aborted by
		// the context, so no extra goroutine is required.
		if d.reachedStage(ctx, waitingFor.stage) {
			return true, nil
		}
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("device is not available: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// reachedStage returns true if the device reports the given configuration
// initialization stage.
func (d *Client) reachedStage(ctx context.Context, stage string) bool {
	conf, err := d.Get(ctx, []string{"/device/diagnostic/confInitStages"})
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
