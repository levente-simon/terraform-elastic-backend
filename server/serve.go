package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	config = &Config{}
	logger *zap.Logger
)

// Start webserver and serve requests
func ServeHttp(configFilePath string, loggerArg *zap.Logger) error {
	logger = loggerArg // Assign passed logger

	r := mux.NewRouter()

	err := config.readConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	r.HandleFunc("/state/{project}", basicAuth(stateHandler))

	exitCh := make(chan error, 2) // Channel size of 2 to handle both HTTP and HTTPS errors

	if config.HttpServer.HttpEnabled {
		// If http enabled, start the http server
		go func() {
			logger.Info("HTTP Server listening", zap.String("address", config.HttpServer.HttpAddress))
			err := http.ListenAndServe(config.HttpServer.HttpAddress, r)
			exitCh <- fmt.Errorf("HTTP Server Failed: %v", err)
		}()
	}

	if config.HttpServer.HttpsEnabled {
		// If https enabled, start the https server
		go func() {
			logger.Info("HTTPS Server listening", zap.String("address", config.HttpServer.HttpsAddress))
			err := http.ListenAndServeTLS(
				config.HttpServer.HttpsAddress,
				config.HttpServer.TLSCertFile,
				config.HttpServer.TLSKeyFile, r)
			exitCh <- fmt.Errorf("HTTPS Server Failed: %v", err)
		}()
	}

	return <-exitCh
}
