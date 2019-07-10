package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Host     = "0.0.0.0"
	Port     = 8000
	Username = "USER"
	Password = "PASS"
)

func TestBrokerParse(t *testing.T) {
	os.Clearenv()
	os.Setenv("BROKER_HOST", Host)
	os.Setenv("BROKER_PORT", strconv.Itoa(Port))
	os.Setenv("BROKER_USERNAME", Username)
	os.Setenv("BROKER_PASSWORD", Password)

	config, err := ParseBrokerServerConfig()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, Host, config.Host)
	assert.Equal(t, Port, config.Port)
	assert.Equal(t, Username, config.Username)
	assert.Equal(t, Password, config.Password)
}

func TestBrokerDefaults(t *testing.T) {
	os.Clearenv()
	config, err := ParseBrokerServerConfig()
	if !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, err)

	// Ensure all fields were populated with the default values.
	assert.Equal(t, DefaultBrokerHost, config.Host)
	assert.Equal(t, DefaultBrokerPort, config.Port)
	assert.Equal(t, DefaultBrokerUsername, config.Username)
	assert.Equal(t, DefaultBrokerPassword, config.Password)
}

func TestInvalidBrokerPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("BROKER_PORT", "NON-NUMERIC")

	_, err := ParseBrokerServerConfig()
	assert.Error(t, err, "Expected error when parsing port number")
}
