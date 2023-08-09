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

func (device *Client) GetSchema() (string, error) {
	client, err := xmlrpc.NewClient(device.url)
	if err != nil {
		return "", err
	}
	defer client.Close()

	result := &struct {
		JSON string
	}{}

	if err := client.Call("getSchema", nil, result); err != nil {
		return "", err
	}
	return result.JSON, nil
}
