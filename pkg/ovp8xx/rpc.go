package ovp8xx

import (
	"alexejk.io/go-xmlrpc"
)

func (device *Client) Get(pointers []string) (Config, error) {
	client, _ := xmlrpc.NewClient(device.url)
	defer client.Close()

	result := &struct {
		JSON string
	}{}

	arg := &struct {
		Pointers []string
	}{Pointers: pointers}

	err := client.Call("get", arg, result)
	if err != nil {
		return *NewConfig(), err
	}

	return *NewConfig(WitJSONString(result.JSON)), nil
}

func (device *Client) Set(conf Config) error {
	client, _ := xmlrpc.NewClient(device.url)
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
	var err error = nil
	client, _ := xmlrpc.NewClient(device.url)
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
	client, _ := xmlrpc.NewClient(device.url)
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
