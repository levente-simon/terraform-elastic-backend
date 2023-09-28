package elasticop

import (
	"context"
	"regexp"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

// Elastic represents a structured Elasticsearch client,
// encapsulating both the native Elasticsearch client and
// additional configuration properties specific to your application.
type Elastic struct {
	// Client represents the native Elasticsearch client.
	Client *elasticsearch.Client

	// Ctx is the context associated with Elasticsearch operations.
	Ctx context.Context

	// Addresses specifies the Elasticsearch cluster addresses.
	Addresses []string `vault:"addresses" default:"https://localhost:9200"`

	// Username for Elasticsearch authentication.
	Username string `vault:"username" default:"elastic"`

	// Password for Elasticsearch authentication.
	Password string `vault:"password" default:"elastic"`

	// StateIndex represents the index name where terraform states are stored.
	StateIndex string `vault:"state_index" default:"terraform-state"`

	// ResourceIndex represents the index name where terraform resources are stored.
	ResourceIndex string `vault:"resource_index" default:"terraform-resources"`

	// LockIndex represents the index name where terraform locks are stored.
	LockIndex string `vault:"lock_index" default:"terraform-locks"`

	// CloudID is the identifier for Elastic Cloud deployments.
	CloudID string `vault:"cloud_id"`

	// ServiceToken is an Elasticsearch service token.
	ServiceToken string `vault:"service_token"`

	// APIKey is the Elasticsearch access key.
	APIKey string `vault:"api_key"`

	// CertificateFingerprint represents the fingerprint for the Elasticsearch certificate.
	CertificateFingerprint string `vault:"certificate_fingerprint"`

	// CaCert specifies the path to the Certificate Authority certificate for Elasticsearch.
	CaCert string

	// Project denotes the specific project or context.
	Project string

	// Encrypt contains compiled regex patterns used to determine which fields to encrypt.
	Encrypt []*regexp.Regexp

	// Logger is the logger instance.
	Logger *zap.Logger
}
