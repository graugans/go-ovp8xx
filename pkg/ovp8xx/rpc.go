package ovp8xx

import (
	"fmt"

	"alexejk.io/go-xmlrpc"
)

func (device *Client) Get(pointers []string) (Config, error) {
	client, _ := xmlrpc.NewClient(fmt.Sprintf("http://%s/api/rpc/v1/com.ifm.efector/", device.host))
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
