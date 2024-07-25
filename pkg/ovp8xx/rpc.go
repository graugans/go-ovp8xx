package ovp8xx

import (
	"errors"
	"net"

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

// Get retrieves the configuration for the specified pointers from the OVP8xx device.
// The pointers parameter is a slice of strings that contains the pointers to retrieve the configuration for.
// The function returns the retrieved configuration as a Config struct and an error if any occurred.
// Example usage:
//
//	config, err := device.Get([]string{"/device", "/ports"})
//	// ...
func (device *Client) Get(pointers []string) (Config, error) {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return *NewConfig(), err
	}
	defer client.Close()

	result := &struct {
		JSON string
	}{}

	arg := &struct {
		Pointers []string
	}{Pointers: pointers}

	if err = client.Call("get", arg, result); err != nil {
		return *NewConfig(), err
	}

	return *NewConfig(WitJSONString(result.JSON)), nil
}

// Set sets the configuration of the OVP8xx device.
// It takes a Config object as input and returns an error if any.
func (device *Client) Set(conf Config) error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()

	arg := &struct {
		Data string
	}{Data: conf.String()}

	return client.Call("set", arg, nil)
}

// GetInit retrieves the initial configuration from the OVP8xx device.
// It returns a Config struct representing the device configuration and an error if any.
func (device *Client) GetInit() (Config, error) {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return *NewConfig(), err
	}
	defer client.Close()

	result := &struct {
		JSON string
	}{}

	if err = client.Call("getInit", nil, result); err != nil {
		return *NewConfig(), err
	}

	return *NewConfig(WitJSONString(result.JSON)), nil
}

// SaveInit saves the configuration of the OVP8xx device.
// If no pointers are provided, it saves the complete configuration.
// If pointers are provided, it saves only the specified configuration pointers.
// It returns an error if there was a problem saving the configuration.
func (device *Client) SaveInit(pointers []string) error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()

	// In case no pointer is given save the complete configuration
	if len(pointers) == 0 {
		return client.Call("saveInit", nil, nil)
	}

	arg := &struct {
		Pointers []string
	}{
		Pointers: pointers,
	}
	return client.Call("saveInit", arg, nil)
}

// FactoryReset performs a factory reset on the OVP8xx device.
// If keepNetworkSettings is set to true, the network settings will be preserved after the reset.
// Returns an error if the factory reset fails or if there is an issue with the XML-RPC client.
func (device *Client) FactoryReset(keepNetworkSettings bool) error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()

	arg := &struct {
		KeepNetworkSettings bool
	}{
		KeepNetworkSettings: keepNetworkSettings,
	}
	return client.Call("factoryReset", arg, nil)
}

// GetSchema retrieves the schema for the specified pointers from the OVP8xx device.
// It returns the schema in JSON format as a string.
// If an error occurs during the retrieval process, it returns an empty string and the error.
func (device *Client) GetSchema(pointers []string) (string, error) {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return "", err
	}
	defer client.Close()

	result := &struct {
		JSON string
	}{}
	arg := &struct {
		Pointers []string
	}{Pointers: pointers}
	if err := client.Call("getSchema", arg, result); err != nil {
		return "", err
	}
	return result.JSON, nil
}

// Reboot sends a reboot command to the OVP8xx device.
// If an error occurs during the connection or the method call, it is returned.
func (device *Client) Reboot() error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()

	return client.Call("reboot", nil, nil)
}

// RebootToSWUpdate reboots the OVP8xx device into software update mode.
// It establishes a connection with the device using XML-RPC and calls the "rebootToRecovery" method.
// This method is typically used to initiate a firmware update on the device.
// Returns an error if there was a problem establishing the connection or calling the method.
func (device *Client) RebootToSWUpdate() error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Call("rebootToRecovery", nil, nil)
}

// Remove does remove an element from the device temporary configuration.
// The scope of this method is limited to the following regular expressions:
// - ^\/applications\/instances\/app\d+$
// - ^\/device\/log\/components\/[a-zA-Z0-9\-_]+$
// - ^\/applications\/instances\/app\d+/presets/\d+$
//
// If an error occurs during the connection or the method call, it is returned.
func (device *Client) Remove(pointer string) error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()

	arg := &struct {
		Pointer string
	}{Pointer: pointer}

	return client.Call("remove", arg, nil)
}

// GetFiltered retrieves a filtered configuration from the DiagnosisClient.
// It takes a Config object as input and returns a Config object and an error.
func (device *DiagnosisClient) GetFiltered(conf Config) (Config, error) {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return *NewConfig(), err
	}
	defer client.Close()

	arg := &struct {
		Data string
	}{Data: conf.String()}

	result := &struct {
		JSON string
	}{}

	if err = client.Call("getFiltered", arg, result); err != nil {
		return *NewConfig(), err
	}

	return *NewConfig(WitJSONString(result.JSON)), nil
}

// GetFilterSchema retrieves the filter schema from the DiagnosisClient.
// It returns a Config object representing the filter schema and an error if any.
func (device *DiagnosisClient) GetFilterSchema() (Config, error) {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return *NewConfig(), err
	}
	defer client.Close()

	result := &struct {
		JSON string
	}{}

	if err = client.Call("getFilterSchema", nil, result); err != nil {
		return *NewConfig(), err
	}

	return *NewConfig(WitJSONString(result.JSON)), nil
}
