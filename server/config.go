package server

import (
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// Config structure defines the configuration schema.
type Config struct {
	// Configuration for Elasticsearch.
	Elasticsearch struct {
		CaCertPath string `yaml:"ca_cert_path"`
	} `yaml:"elasticsearch"`

	// Configuration for HTTP/HTTPS servers.
	HttpServer struct {
		HttpEnabled  bool   `yaml:"http_enabled"`
		HttpAddress  string `yaml:"http_address"`
		HttpsEnabled bool   `yaml:"https_enabled"`
		HttpsAddress string `yaml:"https_address"`
		TLSCertFile  string `yaml:"tls_cert_file"`
		TLSKeyFile   string `yaml:"tls_key_file"`
	} `yaml:"http_server"`

	// Configuration for Vault.
	Vault struct {
		Address      string `yaml:"address"`
		CACertPath   string `yaml:"ca_cert_path"`
		UserPassPath string `yaml:"userpass_path"`
		KvMountPath  string `yaml:"kv_mount_path"`
		TransitPath  string `yaml:"transit_path"`
	} `yaml:"vault"`

	// List of fields or configurations to encrypt.
	Encrypt []string `yaml:"encrypt"`
}

// setDefaultValues initializes the configuration with default values.
func (c *Config) setDefaultValues() {
	c.HttpServer.HttpEnabled = true
	c.HttpServer.HttpAddress = ":8080"
	c.HttpServer.HttpsEnabled = false
	c.HttpServer.HttpsAddress = ":8443"
	c.HttpServer.TLSCertFile = "cert.pem"
	c.HttpServer.TLSKeyFile = "key.pem"
	c.Vault.Address = "http://localhost:8200"
	c.Vault.UserPassPath = "userpass"
	c.Vault.TransitPath = "transit"
	c.Vault.KvMountPath = "kv"
}

// readConfig reads the configuration from a given file path.
// If the file doesn't exist, it sets the default values.
func (c *Config) readConfig(configFilePath string) error {
	// Assign default values.
	c.setDefaultValues()

	// Attempt to read the configuration file.
	yamlFile, err := os.ReadFile(configFilePath)
	if err != nil {
		// If file doesn't exist, use the default values.
		if os.IsNotExist(err) {
			logger.Warn("Configuration file not found, using default values.")
			return nil
		}
		// If another error occurred, return it.
		return err
	}

	// Unmarshal the YAML content into the Config structure.
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}

	// Log successful configuration load.
	logger.Info("Configuration loaded from file", zap.String("path", configFilePath))
	return nil
}
