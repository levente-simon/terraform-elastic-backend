package vaultop

import (
	vault "github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// contextKey is a custom type used to define keys for context values.
type contextKey string

// VaultClientKey is a context key used to store and retrieve the Vault client from context.
const (
	VaultClientKey contextKey = "vaultClient"
)

// Vault struct represents a Vault client and contains fields for configuration,
// paths, and a zap logger for logging activities related to Vault operations.
type Vault struct {
	// Client is the instance of the Vault client.
	Client *vault.Client

	// Address is the URL where the Vault server is accessible.
	Address string

	// CA path of vault
	CaCertPath string

	// Enable Insecure HTTPS communication
	Insecure bool

	// UserPassPath is the path to the userpass authentication backend in Vault.
	UserPassPath string

	// KvMountPath is the mount path for the configuration.
	KvMountPath string

	// TransitPath is the path for the transit secret engine.
	TransitPath string

	// Logger is the zap logger instance for logging Vault-related operations.
	Logger *zap.Logger
}
