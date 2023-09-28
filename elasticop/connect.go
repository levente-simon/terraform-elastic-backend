package elasticop

import (
	"context"
	"net/url"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/levente-simon/terraform-elastic-backend/vaultop"
	"go.uber.org/zap"
)

// ConnectCluster establishes a connection to the Elasticsearch cluster using the provided configuration
// and populates the Elastic struct's Client with the resulting client.
func (e *Elastic) ConnectCluster(ctx context.Context) error {

	var cert []byte
	var err error

	// Fetch the configuration for the specified project from Vault.
	err = ctx.Value(vaultop.VaultClientKey).(*vaultop.Vault).GetConfig(e.Project, e)
	if err != nil {
		e.Logger.Error("Failed to fetch configuration from Vault", zap.String("project", e.Project), zap.Error(err))
		return err
	}
	e.Logger.Info("Successfully fetched configuration from Vault", zap.String("project", e.Project))

	// If the scheme for any of the addresses is https://, then read the CA certificate.
	for _, address := range e.Addresses {
		parsedURL, err := url.Parse(address)
		if err != nil {
			e.Logger.Error("Failed to parse Elastic address URL", zap.Error(err))
			return err
		}

		if parsedURL.Scheme == "https" {
			cert, err = os.ReadFile(e.CaCert)
			if err != nil {
				e.Logger.Error("Failed to read CA certificate", zap.Error(err))
				return err
			}
			break
		}
	}

	// Define the Elasticsearch configuration based on the populated Elastic struct.
	cfg := elasticsearch.Config{
		Addresses:              e.Addresses,
		Username:               e.Username,
		Password:               e.Password,
		CACert:                 cert,
		CloudID:                e.CloudID,
		ServiceToken:           e.ServiceToken,
		APIKey:                 e.APIKey,
		CertificateFingerprint: e.CertificateFingerprint,
	}

	// Create a new Elasticsearch client using the defined configuration.
	e.Client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		e.Logger.Error("Failed to initialize Elasticsearch client", zap.Error(err))
		return err
	}
	e.Logger.Info("Successfully initialized Elasticsearch client", zap.Strings("addresses", e.Addresses))

	// Store the context for further use.
	e.Ctx = ctx

	return nil
}
