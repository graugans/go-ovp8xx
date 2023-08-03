package config

import (
	"fmt"

	"alexejk.io/go-xmlrpc"
)

type OVP8xx struct {
	Host string
}

func (device *OVP8xx) Get(pointers []string) (string, error) {
	client, _ := xmlrpc.NewClient(fmt.Sprintf("http://%s/api/rpc/v1/com.ifm.efector/", device.Host))
	defer client.Close()

	result := &struct {
		JSON string
	}{}
	arg := &struct {
		Pointers []string
	}{Pointers: pointers}
	err := client.Call("get", arg, result)
	if err != nil {
		return "", err

	}

	return result.JSON, nil
}
