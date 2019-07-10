package config

// These constant are the default values used in case some variable is missing.
const (
	DefaultBrokerHost     = "127.0.0.1"
	DefaultBrokerPort     = 4000
	DefaultBrokerUsername = "username"
	DefaultBrokerPassword = "password"
)

// BrokerServerConfig contains the different configuration settings available for
// the broker API server.
type BrokerServerConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// ParseBrokerServerConfig will parse environment variables and generate a
// BrokerServerConfig.
func ParseBrokerServerConfig() (*BrokerServerConfig, error) {
	host := getVarWithDefault("BROKER_HOST", DefaultBrokerHost)

	port, err := getIntVarWithDefault("BROKER_PORT", DefaultBrokerPort)
	if err != nil {
		return nil, err
	}

	username := getVarWithDefault("BROKER_USERNAME", DefaultBrokerUsername)
	password := getVarWithDefault("BROKER_PASSWORD", DefaultBrokerPassword)

	return &BrokerServerConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}, nil
}
