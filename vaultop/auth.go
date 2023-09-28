package vaultop

import (
	"fmt"
	"net/url"

	vault "github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// BasicAuth authenticates a user against Vault using the userpass authentication backend.
// It takes in a username, password, and the path to the userpass backend in Vault.
// On successful authentication, the client token is set in the Vault client.
// The method returns true on successful authentication, and false otherwise.
func (v *Vault) BasicAuth(username, password, pathUserPass string) (bool, error) {
	v.Logger.Info("Attempting to authenticate user", zap.String("username", username))

	// check the Vault URL
	parsedURL, err := url.Parse(v.Address)
	if err != nil {
		v.Logger.Error("Failed to parse Vault address URL", zap.Error(err))
		return false, err
	}

	// Construct the configuration for the Vault client.
	vaultConfig := &vault.Config{
		Address: v.Address,
	}

	// Configure TLS if the url scheme is https
	if parsedURL.Scheme == "https" {
		tlsConfig := &vault.TLSConfig{
			CACert:   v.CaCertPath,
			Insecure: v.Insecure,
		}
		err = vaultConfig.ConfigureTLS(tlsConfig)
		if err != nil {
			v.Logger.Error("Failed to configure TLS", zap.Error(err))
			return false, err
		}
	}

	// Initialize a new Vault client.
	v.Client, err = vault.NewClient(vaultConfig)
	if err != nil {
		v.Logger.Error("Failed to initialize Vault client", zap.Error(err))
		return false, err
	}

	// Prepare data for authentication.
	data := map[string]interface{}{
		"password": password,
	}

	// Construct the path for userpass authentication.
	path := fmt.Sprintf("auth/%s/login/%s", pathUserPass, username)

	// Attempt to authenticate.
	secret, err := v.Client.Logical().Write(path, data)
	if err != nil || secret == nil {
		v.Logger.Error("Authentication failed", zap.Error(err))
		return false, err
	}

	// Set the client token obtained from the successful authentication.
	v.Client.SetToken(secret.Auth.ClientToken)

	v.Logger.Info("Successfully authenticated user", zap.String("username", username))
	return true, nil
}
