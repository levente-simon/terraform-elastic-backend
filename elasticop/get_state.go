package elasticop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// GetState retrieves the latest state stored in Elasticsearch, which includes the state itself
// and associated resources. It returns the combined state as a JSON byte slice.
func (e *Elastic) GetState() ([]byte, int, error) {

	var buf bytes.Buffer

	// Define Elasticsearch query to fetch the latest state based on the timestamp.
	query := map[string]interface{}{
		"size": 1,
		"sort": []map[string]interface{}{
			{
				"timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}
	// Encode the Elasticsearch query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		e.Logger.Error("Error encoding Elasticsearch query", zap.Error(err))
		return nil, http.StatusInternalServerError, fmt.Errorf("error encoding query: %s", err)
	}

	// Search Elasticsearch for the state data
	res, err := e.Client.Search(
		e.Client.Search.WithContext(e.Ctx),
		e.Client.Search.WithIndex(e.StateIndex),
		e.Client.Search.WithBody(&buf),
		e.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		// Error getting the respose
		e.Logger.Error("Error getting Elasticsearch response", zap.Error(err))
		return nil, http.StatusInternalServerError, fmt.Errorf("error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		e.Logger.Error("Error finding the document in Elasticsearch", zap.Error(err))
		return nil, http.StatusNotFound, fmt.Errorf("state not found")
	}

	// Parse response from the Elasticsearch
	var esResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		e.Logger.Error("Error parsing the response from Elasticsearch", zap.Error(err))
		return nil, http.StatusInternalServerError, fmt.Errorf("error parsing Elasticsearch response: %s", err)
	}

	// Get the "source" and "timestamp" variables from the result
	var timestamp string
	var source map[string]interface{}
	if hits, ok := esResponse["hits"].(map[string]interface{}); ok {
		if hitList, ok := hits["hits"].([]interface{}); ok && len(hitList) > 0 {
			hitMap := hitList[0].(map[string]interface{})
			source = hitMap["_source"].(map[string]interface{})
			timestamp = source["timestamp"].(string)
		}
	}

	// Get the resources connected to the state
	resources, err := e.GetResources(timestamp)
	if err != nil {
		e.Logger.Error("Error fetching the resources from Elasticsearch", zap.Error(err))
		return nil, http.StatusInternalServerError, fmt.Errorf("error fetching resources: %s", err)
	}
	source["resources"] = resources

	// Decrypt encrypted fields
	e.TraverseAndModify(source, e.Encrypt, false)

	// Marshal state data to response json
	jsonData, err := json.Marshal(source)
	if err != nil {
		e.Logger.Error("Failed to marshal state data to response", zap.Error(err))
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to marshal state data: %s", err)
	}

	e.Logger.Info("Successfully retrieved state from Elasticsearch", zap.String("timestamp", timestamp), zap.String("project", e.Project))
	return jsonData, http.StatusOK, nil
}

// GetResources retrieves the resources associated with a given timestamp from Elasticsearch.
// It returns the resources as a slice of map[string]interface{}.
func (e *Elastic) GetResources(timestamp string) ([]map[string]interface{}, error) {
	var buf bytes.Buffer

	// Define Elasticsearch query to fetch resources based on the timestamp.
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"timestamp": timestamp,
			},
		},
	}

	// Encode the Elasticsearch query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		e.Logger.Error("Error encoding Elasticsearch query", zap.Error(err))
		return nil, err
	}

	// Search for the resources in Elasticsearch
	res, err := e.Client.Search(
		e.Client.Search.WithContext(e.Ctx),
		e.Client.Search.WithIndex(e.ResourceIndex),
		e.Client.Search.WithBody(&buf),
	)
	if err != nil {
		e.Logger.Error("Error getting Elasticsearch response", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	// Parse response from Elasticsearch
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		e.Logger.Error("Error parsing the response from Elasticsearch", zap.Error(err))
		return nil, err
	}

	var resources []map[string]interface{}
	if hits, ok := r["hits"].(map[string]interface{}); ok {
		if hitList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitList {
				hitMap := hit.(map[string]interface{})
				source := hitMap["_source"].(map[string]interface{})
				resources = append(resources, source)
			}
		}
	}
	return resources, nil
}
