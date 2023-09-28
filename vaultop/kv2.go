package vaultop

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// GetKv2Secret retrieves a KV version 2 secret stored in Vault.
// The function requires the mount path and the secret path in Vault.
// Returns a map containing the secret data or an error if unsuccessful.
func (v *Vault) GetKv2Secret(mountPath string, secretPath string) (map[string]interface{}, error) {
	v.Logger.Info("Fetching KVv2 secret from Vault", zap.String("mountPath", mountPath), zap.String("secretPath", secretPath))

	// Obtain the KVv2 secret engine from the client.
	kv := v.Client.KVv2(mountPath)
	if kv == nil {
		v.Logger.Error("No data found for KVv2 secret", zap.String("mountPath", mountPath), zap.String("secretPath", secretPath))
		return nil, fmt.Errorf("no data found")
	}

	// Attempt to retrieve the secret using the provided secret path.
	secret, err := kv.Get(context.Background(), secretPath)
	if err != nil {
		v.Logger.Error("Error fetching secret from Vault", zap.Error(err))
		return nil, fmt.Errorf("error getting vault config")
	}

	return secret.Data, nil
}

// GetConfig maps configuration data from Vault into the provided 'config' structure.
// It dynamically reads the 'vault' and 'default' struct tags to know where to pull data from Vault
// and where to set default values if the data is missing in Vault.
// The function updates the fields in the 'config' structure in place.
func (v *Vault) GetConfig(secretPath string, config interface{}) error {
	v.Logger.Info("Fetching configuration from Vault", zap.String("secretPath", secretPath))

	// Retrieve the raw configuration data from Vault.
	data, err := v.GetKv2Secret(v.KvMountPath, secretPath)
	if err != nil {
		v.Logger.Error("Failed to retrieve KVv2 secret", zap.String("secretPath", secretPath), zap.Error(err))
		return err
	}

	// Use reflection to dynamically map Vault data to the config struct.
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			// Skip fields that cannot be set.
			continue
		}
		// Check for the 'vault' tag on the struct field.
		if tag, ok := typ.Field(i).Tag.Lookup("vault"); ok {
			var value interface{}
			// Try to retrieve the value from the Vault data based on the 'vault' tag.
			if tagValue, exist := data[tag]; exist {
				value = tagValue
			} else if tag, ok := typ.Field(i).Tag.Lookup("default"); ok {
				// If the value doesn't exist in Vault, try to set a default value based on the 'default' tag.
				value = tag
			} else {
				continue
			}

			// Depending on the type of the struct field, set its value.
			switch field.Kind() {
			case reflect.String:
				field.SetString(fmt.Sprintf("%v", value))
			case reflect.Slice:
				str, ok := value.(string)
				if ok {
					field.Set(reflect.ValueOf(strings.Split(str, ",")))
				}
			case reflect.Int:
				str, ok := value.(string)
				if ok {
					intVal, err := strconv.Atoi(str)
					if err != nil {
						v.Logger.Error("Error converting value to integer", zap.String("value", str), zap.Error(err))
						return err
					}
					field.SetInt(int64(intVal))
				}
			}
		}
	}
	v.Logger.Info("Configuration successfully populated from Vault", zap.String("secretPath", secretPath))
	return nil
}
