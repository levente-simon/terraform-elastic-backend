package elasticop

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/levente-simon/terraform-elastic-backend/vaultop"
	"go.uber.org/zap"
)

// TraverseAndModify is a recursive function that traverses through the node's structure (which can be maps or slices).
// Depending on the 'encrypt' flag, it either encrypts or decrypts the relevant fields.
// Encryption is based on matching regex patterns. Decryption is based on value prefixes.
func (e *Elastic) TraverseAndModify(node interface{}, compiledRegex []*regexp.Regexp, encrypt bool, paths ...string) error {
	var currentPath string
	if len(paths) > 0 {
		currentPath = paths[0]
	} else {
		currentPath = ""
	}

	switch v := node.(type) {
	case map[string]interface{}: // For JSON objects
		for k, val := range v {
			newPath := currentPath + "." + k

			if encrypt { // Encryption based on regex pattern matching
				for _, re := range compiledRegex {
					if re.MatchString(newPath) {
						encryptedVal, err := e.Ctx.Value(vaultop.VaultClientKey).(*vaultop.Vault).EncryptWithVault(fmt.Sprint(val), e.Project)
						if err != nil {
							e.Logger.Error("Failed to encrypt using vault", zap.String("path", newPath), zap.Error(err))
							return err
						}
						v[k] = "tfb_" + encryptedVal
						break
					}
				}
			} else { // Decryption based on specific value prefix
				strVal, isString := val.(string)
				if isString && strings.HasPrefix(strVal, "tfb_vault:") {
					decryptedVal, err := e.Ctx.Value(vaultop.VaultClientKey).(*vaultop.Vault).DecryptWithVault(strings.TrimPrefix(strVal, "tfb_"), e.Project)
					if err != nil {
						e.Logger.Error("Failed to decrypt using vault", zap.String("path", newPath), zap.Error(err))
						return err
					}
					v[k] = decryptedVal
				}
			}

			// Continue traversal recursively to handle nested structures
			if err := e.TraverseAndModify(val, compiledRegex, encrypt, newPath); err != nil {
				return err
			}
		}

	case []interface{}: // For JSON arrays
		for i, val := range v {
			newPath := currentPath + "[" + strconv.Itoa(i) + "]"

			// Encryption and decryption logic similar to the map handling above
			if encrypt {
				for _, re := range compiledRegex {
					if re.MatchString(newPath) {
						encryptedVal, err := e.Ctx.Value(vaultop.VaultClientKey).(*vaultop.Vault).EncryptWithVault(fmt.Sprint(val), e.Project)
						if err != nil {
							e.Logger.Error("Failed to encrypt array item using vault", zap.String("path", newPath), zap.Error(err))
							return err
						}
						v[i] = "tfb_" + encryptedVal
						break
					}
				}
			} else {
				strVal, isString := val.(string)
				if isString && strings.HasPrefix(strVal, "tfb_vault:") {
					decryptedVal, err := e.Ctx.Value(vaultop.VaultClientKey).(*vaultop.Vault).DecryptWithVault(strings.TrimPrefix(strVal, "tfb_"), e.Project)
					if err != nil {
						e.Logger.Error("Failed to decrypt array item using vault", zap.String("path", newPath), zap.Error(err))
						return err
					}
					v[i] = decryptedVal
				}
			}

			// Continue traversal recursively to handle nested structures
			if err := e.TraverseAndModify(val, compiledRegex, encrypt, newPath); err != nil {
				return err
			}
		}
	}

	return nil
}
