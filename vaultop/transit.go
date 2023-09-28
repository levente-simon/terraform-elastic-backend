package vaultop

import (
	"encoding/base64"
	"fmt"

	"go.uber.org/zap"
)

// EncryptWithVault uses Vault's Transit secret engine to encrypt the given value.
// The function requires a key name to perform the encryption.
// It returns the encrypted ciphertext or an error if unsuccessful.
func (v *Vault) EncryptWithVault(value, key string) (string, error) {
	// Encode the input value into base64 format.
	encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
	data := map[string]interface{}{
		"plaintext": encodedValue,
	}

	// Write the plaintext data to Vault's Transit secret engine for encryption.
	secret, err := v.Client.Logical().Write(v.TransitPath+"/encrypt/"+key, data)
	if err != nil {
		v.Logger.Error("Error encrypting data with Vault", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("error encrypting data with Vault: %v", err)
	}

	// Extract the ciphertext from Vault's response.
	ciphertext, ok := secret.Data["ciphertext"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get ciphertext from Vault response")
	}

	return ciphertext, nil
}

// DecryptWithVault uses Vault's Transit secret engine to decrypt the given ciphertext.
// The function requires a key name to perform the decryption.
// It returns the decrypted plaintext or an error if unsuccessful.
func (v *Vault) DecryptWithVault(ciphertext, key string) (string, error) {
	data := map[string]interface{}{
		"ciphertext": ciphertext,
	}

	// Write the ciphertext data to Vault's Transit secret engine for decryption.
	secret, err := v.Client.Logical().Write(v.TransitPath+"/decrypt/"+key, data)
	if err != nil {
		v.Logger.Error("Error decrypting data with Vault", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("error decrypting data with Vault: %v", err)
	}

	// Extract the plaintext from Vault's response.
	plaintext, ok := secret.Data["plaintext"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get plaintext from Vault response")
	}

	// Decode the base64 encoded plaintext.
	decodedValue, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		v.Logger.Error("Failed to decode the secret from Vault", zap.Error(err))
		return "", fmt.Errorf("failed to decode the secret in vault: %v", err)
	}

	return string(decodedValue), nil
}
