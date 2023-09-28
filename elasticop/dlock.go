package elasticop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Define the index in Elasticsearch where the locks will be stored.
const lockIndex = "terraform-locks"

// ESDistributedLock represents the structure of a lock stored in Elasticsearch.
type ESDistributedLock struct {
	Project  string `json:"project"`
	LockedBy string `json:"lockedBy"`
	Version  int64  `json:"version"`
}

// AcquireLock attempts to create a lock for the specified project.
// If the lock already exists, it returns false, otherwise it returns true.
func (e *Elastic) AcquireLock(project, remoteId string) (bool, error) {
	var buf bytes.Buffer
	lock := ESDistributedLock{
		Project:  project,
		LockedBy: remoteId,
	}

	// Convert the lock into JSON format.
	if err := json.NewEncoder(&buf).Encode(lock); err != nil {
		e.Logger.Error("Error encoding lock data", zap.Error(err))
		return false, err
	}

	// Attempt to create the lock in Elasticsearch.
	res, err := e.Client.Index(
		lockIndex,
		&buf,
		e.Client.Index.WithDocumentID(project),
		e.Client.Index.WithRefresh("true"),
	)
	defer res.Body.Close()

	if err != nil {
		// Check if the error is because the lock already exists.
		if res.StatusCode == http.StatusConflict {
			e.Logger.Info("Lock already exists", zap.String("project", project))
			return false, nil
		}
		e.Logger.Error("Error creating the lock in Elasticsearch", zap.Error(err))
		return false, err
	}

	// If the lock was successfully created, return true.
	if res.StatusCode == http.StatusCreated {
		e.Logger.Info("Lock successfully acquired", zap.String("project", project))
		return true, nil
	}

	e.Logger.Warn("Failed to acquire lock, unknown reason", zap.Int("status_code", res.StatusCode))
	return false, nil
}

// ReleaseLock attempts to release a lock for the specified project.
// Returns true if the lock was successfully released or if the lock didn't exist, false otherwise.
func (e *Elastic) ReleaseLock(project string) (bool, error) {
	// Attempt to delete the lock in Elasticsearch.
	res, err := e.Client.Delete(
		lockIndex,
		project,
		e.Client.Delete.WithRefresh("true"),
	)
	defer res.Body.Close()

	if err != nil {
		e.Logger.Error("Error deleting the lock", zap.Error(err))
		return false, err
	}

	// If the lock was successfully deleted or didn't exist, return true.
	if res.StatusCode == http.StatusOK {
		e.Logger.Info("Lock successfully released", zap.String("project", project))
		return true, nil
	} else if res.StatusCode == http.StatusNotFound {
		e.Logger.Info("Lock did not exist", zap.String("project", project))
		return false, nil
	}

	e.Logger.Warn("Failed to release lock, unknown reason", zap.Int("status_code", res.StatusCode))
	return false, fmt.Errorf("unexpected status code: %d", res.StatusCode)
}
