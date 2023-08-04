package ovp8xx

type (
	ConfigOption func(c *Config)
	Config       struct {
		JSON string
	}
)

func NewConfig(opts ...ConfigOption) *Config {
	// Initialise with default values
	config := &Config{
		JSON: "{}",
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	return config
}

func WitJSONString(json string) ConfigOption {
	return func(c *Config) {
		c.JSON = json
	}
}

func (c Config) String() string {
	return c.JSON
}
