package main

import (
	"fmt"
	"net/http"
	"strconv"

	"os"

	"strings"

	"code.cloudfoundry.org/lager"
	atlasclient "github.com/10gen/atlas-service-broker/pkg/atlas"
	atlasbroker "github.com/10gen/atlas-service-broker/pkg/broker"
	"github.com/pivotal-cf/brokerapi"
	"go.uber.org/zap"
)

// Default values for the configuration variables.
const (
	DefaultAtlasBaseURL = "https://cloud.mongodb.com"
	DefaultServerHost   = "127.0.0.1"
	DefaultServerPort   = 4000
)

func main() {
	// Zap is the main logger used for the broker. A lager logger is also
	// created as the brokerapi library requires one.
	zapLogger, _ := zap.NewDevelopment()
	logger := zapLogger.Sugar()

	// Try parsing Atlas client config.
	baseURL := strings.TrimRight(getEnvOrDefault("ATLAS_BASE_URL", DefaultAtlasBaseURL), "/")
	groupID := getEnvOrPanic("ATLAS_GROUP_ID")
	publicKey := getEnvOrPanic("ATLAS_PUBLIC_KEY")
	privateKey := getEnvOrPanic("ATLAS_PRIVATE_KEY")

	client, err := atlasclient.NewClient(baseURL, groupID, publicKey, privateKey)
	if err != nil {
		logger.Fatal(err)
	}

	// Create broker with the previously created Atlas client.
	broker := atlasbroker.NewBroker(client, logger)

	// Try parsing server config and set up broker API server.
	username := getEnvOrPanic("BROKER_USERNAME")
	password := getEnvOrPanic("BROKER_PASSWORD")
	host := getEnvOrDefault("BROKER_HOST", DefaultServerHost)
	port := getIntEnvOrDefault("BROKER_PORT", DefaultServerPort)

	credentials := brokerapi.BrokerCredentials{
		Username: username,
		Password: password,
	}

	endpoint := host + ":" + strconv.Itoa(port)

	// Mount broker server at the root.
	http.Handle("/", brokerapi.New(broker, lager.NewLogger("api"), credentials))

	logger.Infow("Starting API server", "host", host, "port", port, "atlas_base_url", baseURL, "atlas_group_id", groupID)

	// Start broker HTTP server.
	if err = http.ListenAndServe(endpoint, nil); err != nil {
		logger.Fatal(err)
	}
}

// getEnvOrPanic will try getting an environment variable and fail with a
// helpful error message in case it doesn't exist.
func getEnvOrPanic(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		panic(fmt.Sprintf(`Could not find environment variable "%s"`, name))
	}

	return value
}

// getEnvOrPanic will try getting an environment variable and return a default
// value in case it doesn't exist.
func getEnvOrDefault(name string, def string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return def
	}

	return value
}

// getIntEnvOrDefault will try getting an environment variable and parse it as
// an integer. In case the variable is not set it will return the default value.
func getIntEnvOrDefault(name string, def int) int {
	value, exists := os.LookupEnv(name)
	if !exists {
		return def
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf(`Environment variable "%s" is not an integer`, name))
	}

	return intValue
}
