package server

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/levente-simon/terraform-elastic-backend/vaultop"
	"go.uber.org/zap"
)

// basicAuth is a middleware that wraps the provided http.HandlerFunc with Basic Authentication
// using Vault to verify the credentials.
func basicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Initialize the vault client outside of the function
		var vaultClient = &vaultop.Vault{
			Address:     config.Vault.Address,
			KvMountPath: config.Vault.KvMountPath,
			TransitPath: config.Vault.TransitPath,
			Logger:      logger,
		}

		// Retrieve the Authorization header value
		auth := r.Header.Get("Authorization")
		if auth == "" {
			logger.Warn("Authorization header missing")
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		// Decode the Basic Authentication payload
		payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if err != nil {
			logger.Warn("Invalid authorization")
			http.Error(w, "Invalid authorization", http.StatusUnauthorized)
			return
		}

		// Split the payload into username and password
		authData := strings.SplitN(string(payload), ":", 2)
		if len(authData) < 2 {
			logger.Warn("Invalid authorization")
			http.Error(w, "Invalid authorization", http.StatusUnauthorized)
			return
		}

		// Verify the credentials using Vault
		isAuthenticated, err := vaultClient.BasicAuth(authData[0], authData[1], config.Vault.UserPassPath)
		if err != nil || !isAuthenticated {
			logger.Warn("Invalid credentials provided", zap.String("user", authData[0]))
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		logger.Info("Authorized request", zap.String("user", authData[0]), zap.String("remote_addr", r.RemoteAddr))

		// Add the vault client to the request context and invoke the original handler
		ctx := context.WithValue(r.Context(), vaultop.VaultClientKey, vaultClient)
		handler(w, r.WithContext(ctx))
	}
}
