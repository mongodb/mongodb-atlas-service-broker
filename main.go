package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"os"

	"github.com/gorilla/mux"
	atlasbroker "github.com/mongodb/mongodb-atlas-service-broker/pkg/broker"
	"github.com/pivotal-cf/brokerapi"
)

// releaseVersion should be set by the linker at compile time.
var releaseVersion = "development-build"

// Default values for the configuration variables.
const (
	DefaultLogLevel = "INFO"

	DefaultAtlasBaseURL = "https://cloud.mongodb.com"

	DefaultServerHost = "127.0.0.1"
	DefaultServerPort = 4000
)

func main() {
	// Add --help and -h flag.
	helpDescription := "Print information about the MongoDB Atlas Service Broker and helpful links."
	helpFlag := flag.Bool("help", false, helpDescription)
	flag.BoolVar(helpFlag, "h", false, helpDescription)

	// Add --version and -v flag.
	versionDescription := "Print current version of MongoDB Atlas Service Broker."
	versionFlag := flag.Bool("version", false, versionDescription)
	flag.BoolVar(versionFlag, "v", false, versionDescription)

	flag.Parse()

	// Output help message if help flag was specified.
	if *helpFlag {
		fmt.Println(getHelpMessage())
		return
	}

	// Output current version if version flag was specified.
	if *versionFlag {
		fmt.Println(releaseVersion)
		return
	}

	startBrokerServer()
}

func getHelpMessage() string {
	const helpMessage = `MongoDB Atlas Service Broker %s

This is a Service Broker which provides access to MongoDB deployments running
in MongoDB Atlas. It conforms to the Open Service Broker specification and can
be used with any compatible platform, for example the Kubernetes Service Catalog.

For instructions on how to install and use the Service Broker please refer to
the documentation: https://docs.mongodb.com/atlas-open-service-broker

Github: https://github.com/mongodb/mongodb-atlas-service-broker
Docker Image: quay.io/mongodb/mongodb-atlas-service-broker`

	return fmt.Sprintf(helpMessage, releaseVersion)
}

func startBrokerServer() {
	logLevel := getEnvOrDefault("BROKER_LOG_LEVEL", DefaultLogLevel)
	logger, err := createLogger(logLevel)
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flushes buffer, if any

	broker := atlasbroker.NewBroker(logger)

	router := mux.NewRouter()
	brokerapi.AttachRoutes(router, broker, NewLagerZapLogger(logger))

	// The auth middleware will convert basic auth credentials into an Atlas
	// client.
	baseURL := strings.TrimRight(getEnvOrDefault("ATLAS_BASE_URL", DefaultAtlasBaseURL), "/")
	router.Use(atlasbroker.AuthMiddleware(baseURL))

	// Configure TLS from environment variables.
	tlsEnabled, tlsCertFile, tlsKeyFile := getTLSConfig(logger)

	host := getEnvOrDefault("BROKER_HOST", DefaultServerHost)
	port := getIntEnvOrDefault("BROKER_PORT", DefaultServerPort)

	logger.Infow("Starting API server", "releaseVersion", releaseVersion, "host", host, "port", port, "tls_enabled", tlsEnabled, "atlas_base_url", baseURL)

	// Start broker HTTP server.
	address := host + ":" + strconv.Itoa(port)

	var serverErr error
	if tlsEnabled {
		serverErr = http.ListenAndServeTLS(address, tlsCertFile, tlsKeyFile, router)
	} else {
		logger.Warn("TLS is disabled")
		serverErr = http.ListenAndServe(address, router)
	}

	if serverErr != nil {
		logger.Fatal(serverErr)
	}
}

func getTLSConfig(logger *zap.SugaredLogger) (bool, string, string) {
	certFile := getEnvOrDefault("BROKER_TLS_CERT_FILE", "")
	keyFile := getEnvOrDefault("BROKER_TLS_KEY_FILE", "")

	hasCertFile := certFile != ""
	hasKeyFile := keyFile != ""

	// Bail if only one of the cert and key has been provided.
	if (hasCertFile && !hasKeyFile) || (!hasCertFile && hasKeyFile) {
		logger.Fatal("Both a certificate and private key are necessary to enable TLS")
	}

	enabled := hasCertFile && hasKeyFile
	return enabled, certFile, keyFile
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

// createLogger will create a zap sugared logger with the specified log level.
func createLogger(levelName string) (*zap.SugaredLogger, error) {
	levelByName := map[string]zapcore.Level{
		"DEBUG": zapcore.DebugLevel,
		"INFO":  zapcore.InfoLevel,
		"WARN":  zapcore.WarnLevel,
		"ERROR": zapcore.ErrorLevel,
	}

	// Convert log level string to a zap level.
	level, ok := levelByName[levelName]
	if !ok {
		return nil, fmt.Errorf(`invalid log level "%s"`, levelName)
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
