package main

import (
	"net/http"
	"strconv"

	"code.cloudfoundry.org/lager"
	atlasclient "github.com/10gen/atlas-service-broker/pkg/atlas"
	atlasbroker "github.com/10gen/atlas-service-broker/pkg/broker"
	"github.com/10gen/atlas-service-broker/pkg/config"
	"github.com/pivotal-cf/brokerapi"
	"go.uber.org/zap"
)

func main() {
	// Zap is the main logger used for the broker. A lager logger is also
	// created as the brokerapi library requires one.
	zapLogger, _ := zap.NewDevelopment()
	logger := zapLogger.Sugar()

	// Try parsing Atlas client config from env variables.
	atlasConfig, err := config.ParseAtlasConfig()
	if err != nil {
		logger.Fatal(err)
	}

	client, err := atlasclient.NewClient(atlasConfig.BaseURL, atlasConfig.GroupID, atlasConfig.PublicKey, atlasConfig.PrivateKey)
	if err != nil {
		logger.Fatal(err)
	}

	// Create broker with the previously created Atlas client.
	broker := atlasbroker.NewBroker(client, logger)

	// Try parsing server config and set up broker API server.
	serverConfig, err := config.ParseBrokerServerConfig()
	if err != nil {
		logger.Fatal(err)
	}

	credentials := brokerapi.BrokerCredentials{
		Username: serverConfig.Username,
		Password: serverConfig.Password,
	}

	endpoint := serverConfig.Host + ":" + strconv.Itoa(serverConfig.Port)

	// Mount broker server at the root.
	http.Handle("/", brokerapi.New(broker, lager.NewLogger("api"), credentials))

	// Start broker HTTP server.
	logger.Infow("Starting API server", "host", serverConfig.Host, "port", serverConfig.Port, "atlas_base_url", atlasConfig.BaseURL, "atlas_group_id", atlasConfig.GroupID)
	logger.Fatal(http.ListenAndServe(endpoint, nil))
}
