package ovp8xx

import (
	"fmt"
)

type (
	ClientOption func(c *Client)
	Client       struct {
		host string
		url  string
	}
)

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

func WithHost(host string) ClientOption {
	return func(c *Client) {
		c.host = host
	}
}
