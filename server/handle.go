package server

import (
	"io"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/levente-simon/terraform-elastic-backend/elasticop"
	"go.uber.org/zap"
)

// stateHandler is the main handler for managing terraform state in Elasticsearch.
func stateHandler(w http.ResponseWriter, r *http.Request) {

	v := mux.Vars(r)

	// Compile the regular expressions from config for fields to encrypt.
	compiledRegex := make([]*regexp.Regexp, len(config.Encrypt))
	for i, pattern := range config.Encrypt {
		compiledRegex[i] = regexp.MustCompile(pattern)
	}

	// Initialize the Elasticsearch client.
	var elastic = &elasticop.Elastic{
		CaCert:  config.Elasticsearch.CaCertPath,
		Project: v["project"],
		Encrypt: compiledRegex,
		Logger:  logger,
	}

	// Connect to the Elasticsearch cluster.
	err := elastic.ConnectCluster(r.Context())
	if err != nil {
		logger.Error("Failed to initialize Elasticsearch client", zap.Error(err), zap.String("project", v["project"]))
		http.Error(w, "Internal server error: Elasticsearch client is not initialized", http.StatusInternalServerError)
		return
	}

	// Handle different HTTP methods.
	switch r.Method {
	case "GET":
		Get(w, r, elastic)
	case "POST":
		Post(w, r, elastic)
	case "LOCK":
		Lock(w, r, elastic)
	case "UNLOCK":
		Unlock(w, r, elastic)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get retrieves the terraform state from Elasticsearch.
func Get(w http.ResponseWriter, r *http.Request, e *elasticop.Elastic) {
	body, httpStatus, err := e.GetState()
	if err != nil {
		logger.Error("Failed to retrieve state from Elasticsearch", zap.Error(err))
		http.Error(w, err.Error(), httpStatus)
		return
	}

	logger.Info("Successfully retrieved state", zap.String("project", e.Project))
	w.Write(body)
}

// Post updates the terraform state in Elasticsearch.
func Post(w http.ResponseWriter, r *http.Request, e *elasticop.Elastic) {
	updatedState, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read state from the request", zap.Error(err))
		http.Error(w, "Error reading state", http.StatusInternalServerError)
		return
	}

	httpStatus, err := e.StoreState(updatedState)
	if err != nil {
		logger.Error("Failed to store state in Elasticsearch", zap.Error(err))
		http.Error(w, err.Error(), httpStatus)
		return
	}

	logger.Info("Successfully stored state", zap.String("project", e.Project))
	w.WriteHeader(http.StatusOK)
}

// Lock attempts to acquire a distributed lock for the specified project.
func Lock(w http.ResponseWriter, r *http.Request, e *elasticop.Elastic) {
	// instanceID := "unique-instance-id" // replace this with a real, unique instance ID for your application
	acquired, err := e.AcquireLock(e.Project, r.RemoteAddr)
	if err != nil {
		logger.Error("Failed to acquire lock", zap.Error(err))
		http.Error(w, "Failed to acquire lock", http.StatusInternalServerError)
		return
	}
	if !acquired {
		http.Error(w, "State is locked by another instance", http.StatusLocked)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Unlock attempts to release a distributed lock for the specified project.
func Unlock(w http.ResponseWriter, r *http.Request, e *elasticop.Elastic) {
	released, err := e.ReleaseLock(e.Project)
	if err != nil {
		logger.Error("Failed to release lock", zap.Error(err))
		http.Error(w, "Failed to release lock", http.StatusInternalServerError)
		return
	}
	if !released {
		http.Error(w, "Failed to release lock, lock does not exist", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
