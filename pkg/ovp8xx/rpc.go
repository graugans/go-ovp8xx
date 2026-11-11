package ovp8xx

import (
	"context"
	"errors"
	"net"
	"net/http"

	"alexejk.io/go-xmlrpc"
)

// IsTimeoutError checks if the given error is a network timeout error.
// It returns true if the error is a network error and the timeout flag is set, otherwise it returns false.
func IsTimeoutError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}

// jsonResult is the reply of all XML-RPC methods returning a JSON document.
type jsonResult struct {
	JSON string
}

// call opens a connection to the endpoint, invokes method and closes the
// connection again. The request is aborted once ctx is done or the configured
// timeout expires. An error while closing the connection is only reported if
// the call itself succeeded.
func (e endpoint) call(ctx context.Context, method string, args, reply any) (err error) {
	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}
	client, err := xmlrpc.NewClient(e.url, xmlrpc.HttpClient(e.contextClient(ctx)))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	return client.Call(method, args, reply)
}

// contextClient returns a copy of the endpoint's HTTP client whose requests
// are bound to ctx. The XML-RPC package does not support a context per call,
// so the context is injected on the transport level.
func (e endpoint) contextClient(ctx context.Context) *http.Client {
	base := e.httpClient
	if base == nil {
		base = http.DefaultClient
	}
	transport := base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	client := *base
	client.Transport = contextTransport{ctx: ctx, base: transport}
	return &client
}

// contextTransport binds every request to ctx.
type contextTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.base.RoundTrip(req.WithContext(t.ctx))
}

// callConfig invokes method and returns the JSON reply as Config.
func (e endpoint) callConfig(ctx context.Context, method string, args any) (Config, error) {
	result := &jsonResult{}
	if err := e.call(ctx, method, args, result); err != nil {
		return *NewConfig(), err
	}
	return *NewConfig(WitJSONString(result.JSON)), nil
}

// Get retrieves the configuration for the specified pointers from the OVP8xx device.
// The pointers parameter is a slice of strings that contains the pointers to retrieve the configuration for.
// The function returns the retrieved configuration as a Config struct and an error if any occurred.
// Example usage:
//
//	config, err := device.Get(ctx, []string{"/device", "/ports"})
//	// ...
func (device *Client) Get(ctx context.Context, pointers []string) (Config, error) {
	arg := &struct {
		Pointers []string
	}{Pointers: pointers}
	return device.callConfig(ctx, "get", arg)
}

// Set sets the configuration of the OVP8xx device.
// It takes a Config object as input and returns an error if any.
func (device *Client) Set(ctx context.Context, conf Config) error {
	arg := &struct {
		Data string
	}{Data: conf.String()}
	return device.call(ctx, "set", arg, nil)
}

// GetInit retrieves the initial configuration from the OVP8xx device.
// It returns a Config struct representing the device configuration and an error if any.
func (device *Client) GetInit(ctx context.Context) (Config, error) {
	return device.callConfig(ctx, "getInit", nil)
}

// SaveInit saves the configuration of the OVP8xx device.
// If no pointers are provided, it saves the complete configuration.
// If pointers are provided, it saves only the specified configuration pointers.
// It returns an error if there was a problem saving the configuration.
func (device *Client) SaveInit(ctx context.Context, pointers []string) error {
	// In case no pointer is given save the complete configuration
	if len(pointers) == 0 {
		return device.call(ctx, "saveInit", nil, nil)
	}
	arg := &struct {
		Pointers []string
	}{Pointers: pointers}
	return device.call(ctx, "saveInit", arg, nil)
}

// FactoryReset performs a factory reset on the OVP8xx device.
// If keepNetworkSettings is set to true, the network settings will be preserved after the reset.
// Returns an error if the factory reset fails or if there is an issue with the XML-RPC client.
func (device *Client) FactoryReset(ctx context.Context, keepNetworkSettings bool) error {
	arg := &struct {
		KeepNetworkSettings bool
	}{KeepNetworkSettings: keepNetworkSettings}
	return device.call(ctx, "factoryReset", arg, nil)
}

// GetSchema retrieves the schema for the specified pointers from the OVP8xx device.
// It returns the schema in JSON format as a string.
// If an error occurs during the retrieval process, it returns an empty string and the error.
func (device *Client) GetSchema(ctx context.Context, pointers []string) (string, error) {
	arg := &struct {
		Pointers []string
	}{Pointers: pointers}
	result := &jsonResult{}
	if err := device.call(ctx, "getSchema", arg, result); err != nil {
		return "", err
	}
	return result.JSON, nil
}

// Reboot sends a reboot command to the OVP8xx device.
// If an error occurs during the connection or the method call, it is returned.
func (device *Client) Reboot(ctx context.Context) error {
	return device.call(ctx, "reboot", nil, nil)
}

// RebootToSWUpdate reboots the OVP8xx device into software update mode.
// It establishes a connection with the device using XML-RPC and calls the "rebootToRecovery" method.
// This method is typically used to initiate a firmware update on the device.
// Returns an error if there was a problem establishing the connection or calling the method.
func (device *Client) RebootToSWUpdate(ctx context.Context) error {
	return device.call(ctx, "rebootToRecovery", nil, nil)
}

// Remove does remove an element from the device temporary configuration.
// The scope of this method is limited to the following regular expressions:
// - ^\/applications\/instances\/app\d+$
// - ^\/device\/log\/components\/[a-zA-Z0-9\-_]+$
// - ^\/applications\/instances\/app\d+/presets/\d+$
//
// If an error occurs during the connection or the method call, it is returned.
func (device *Client) Remove(ctx context.Context, pointer string) error {
	arg := &struct {
		Pointer string
	}{Pointer: pointer}
	return device.call(ctx, "remove", arg, nil)
}

// GetFiltered retrieves a filtered configuration from the DiagnosisClient.
// It takes a Config object as input and returns a Config object and an error.
func (device *DiagnosisClient) GetFiltered(ctx context.Context, conf Config) (Config, error) {
	arg := &struct {
		Data string
	}{Data: conf.String()}
	return device.callConfig(ctx, "getFiltered", arg)
}

// GetFilterSchema retrieves the filter schema from the DiagnosisClient.
// It returns a Config object representing the filter schema and an error if any.
func (device *DiagnosisClient) GetFilterSchema(ctx context.Context) (Config, error) {
	return device.callConfig(ctx, "getFilterSchema", nil)
}
