package elasticop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// StoreState saves the provided state to Elasticsearch. It first stores individual resources
// and then the entire state minus the resources.
func (e *Elastic) StoreState(updatedState []byte) (int, error) {
	var buf bytes.Buffer

	// Get the current timestamp in UTC format.
	currentTime := time.Now().UTC().Format(time.RFC3339)

	// Parse the input state to an interface type.
	var stateData interface{}
	err := json.Unmarshal([]byte(updatedState), &stateData)
	if err != nil {
		e.Logger.Error("Failed to unmarshal updatedState", zap.Error(err))
		return http.StatusInternalServerError, fmt.Errorf("failed to unmarshal updatedState: %s", err)
	}

	// Modify data if encryption is required.
	e.TraverseAndModify(stateData, e.Encrypt, true)

	// Assert the state data to a map.
	var stateMap map[string]interface{} = stateData.(map[string]interface{})

	// Extract the resources section from the state data.
	resources, ok := stateMap["resources"].([]interface{})
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("malformed state: missing resources")
	}

	// Process and store each resource in Elasticsearch.
	for _, resource := range resources {
		var resourceMap map[string]interface{}
		if resourceMap, ok = resource.(map[string]interface{}); ok {
			resourceMap["timestamp"] = currentTime
		} else {
			return http.StatusInternalServerError, fmt.Errorf("failed to type assert resource to map[string]interface{}")
		}

		// Reset buffer for new data.
		buf.Reset()

		// Encode the resource data.
		if err := json.NewEncoder(&buf).Encode(resourceMap); err != nil {
			e.Logger.Error("Error encoding resource data", zap.Error(err))
			return http.StatusInternalServerError, err
		}

		// Save the resource data to Elasticsearch.
		res, err := e.Client.Index(
			e.ResourceIndex,
			&buf,
			e.Client.Index.WithRefresh("true"),
		)
		if err != nil || res.IsError() {
			e.Logger.Error("Error saving resource to Elasticsearch", zap.Error(err))
			return http.StatusInternalServerError, fmt.Errorf("error saving resource: %s", err)
		}
		defer res.Body.Close()
	}

	// Remove the resources key from the state map since we've already processed them.
	delete(stateMap, "resources")

	// Add a timestamp to the state data.
	stateMap["timestamp"] = currentTime

	// Reset buffer for new data.
	buf.Reset()

	// Encode the state data.
	if err := json.NewEncoder(&buf).Encode(stateMap); err != nil {
		e.Logger.Error("Error Error encoding data", zap.Error(err))
		return http.StatusInternalServerError, err
	}

	// Save the state data to Elasticsearch.
	res, err := e.Client.Index(
		e.StateIndex,
		&buf,
		e.Client.Index.WithRefresh("true"),
	)
	if err != nil || res.IsError() {
		e.Logger.Error("Error saving state to Elasticsearch", zap.Error(err))
		return http.StatusInternalServerError, fmt.Errorf("error saving state: %s", err)
	}
	defer res.Body.Close()

	// Log the successful operation.
	e.Logger.Info("Successfully stored state to Elasticsearch", zap.String("timestamp", currentTime), zap.String("project", e.Project))

	return http.StatusOK, nil
}
