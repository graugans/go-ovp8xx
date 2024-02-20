package ovp8xx

import (
	"alexejk.io/go-xmlrpc"
)

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

func (device *Client) Set(conf Config) error {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return err
	}
	defer client.Close()

	arg := &struct {
		Data string
	}{Data: conf.String()}

	if err := client.Call("set", arg, nil); err != nil {
		return err
	}

	return nil
}

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
	}{Pointers: pointers}
	return client.Call("saveInit", arg, nil)
}

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
