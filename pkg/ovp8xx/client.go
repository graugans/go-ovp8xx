package ovp8xx

type (
	ClientOption func(c *Client)
	Client       struct {
		host string
	}
)

func NewClient(opts ...ClientOption) *Client {
	// Initialise with default values
	client := &Client{
		host: getEnv("OVP8XX_IP", "192.168.0.69"),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

func WithHost(host string) ClientOption {
	return func(c *Client) {
		c.host = host
	}
}
