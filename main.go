package main

import (
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/10gen/atlas-service-broker/pkg/atlas"
	"github.com/10gen/atlas-service-broker/pkg/broker"
	"github.com/pivotal-cf/brokerapi"
	"go.uber.org/zap"
)

func main() {
	zapLogger, _ := zap.NewDevelopment()
	logger := zapLogger.Sugar()

	baseURL := os.Getenv("ATLAS_BASE_URL")
	groupID := os.Getenv("ATLAS_GROUP_ID")
	publicKey := os.Getenv("ATLAS_PUBLIC_KEY")
	privateKey := os.Getenv("ATLAS_PRIVATE_KEY")

	client, err := atlas.NewClient(baseURL, groupID, publicKey, privateKey)
	if err != nil {
		logger.Fatal(err)
	}

	broker := broker.NewBroker(client, logger)
	credentials := brokerapi.BrokerCredentials{
		Username: "username",
		Password: "password",
	}

	api := brokerapi.New(broker, lager.NewLogger("api"), credentials)
	http.Handle("/", api)

	host := os.Getenv("BROKER_HOST")
	port := os.Getenv("BROKER_PORT")
	endpoint := host + ":" + port
	logger.Infof("Starting API server listening on %s", endpoint)
	logger.Fatal(http.ListenAndServe(endpoint, nil))
}
