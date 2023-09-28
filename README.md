## Terraform Elastic Backend

Terraform Elastic Backend is a custom backend server designed to store Terraform state in Elasticsearch. It leverages Vault for secure authentication and data encryption.

### Features:

- **Elasticsearch Integration**: Utilizes Elasticsearch to store Terraform state, offering scalability and resilience.
  
- **Enhanced State Management**: Distinguishes between state and resources by maintaining them as individual documents. Additionally, every change is timestamped, enabling precise change tracking and facilitating advanced reporting or visualization.

- **Distributed Locking**: Employs an Elasticsearch index for state locking, ensuring consistency across distributed deployments.

- **Secure Authentication**: Integrates with Vault's userpass engine for foundational authentication. Expansion to include more authentication methods is underway.

- **Dynamic Configuration**: Adopts a project-centric approach, deriving configuration from the URL and storing it within Vault's KV2 engine. 

- **Robust Authorization**: Fetches configuration (like access credentials to the Elastic cluster and state storage indexes) based on the authenticated user's rights in Vault, ensuring granular access control through Vault's policies.

- **Data Encryption**: Prioritizes data security by encrypting sensitive data (identified via regex patterns) using Vault's Transit backend.

### Usage:

1. Start the server by pointing to an optional configuration file:
   ```
   ./terraform-backend --config path/to/config.yml
   ```

2. By default, the server will start on port 8080 for HTTP and 8443 for HTTPS. You can customize this and other settings in the configuration file.

## Configuration:

Sample configuration (`config.yml`):

```yaml
elasticsearch:
  ca_cert_path: "/path/to/ca/cert"
http_server:
  http_enabled: true
  http_address: ":8080"
  https_enabled: false
  https_address: ":8443"
  tls_cert_file: "cert.pem"
  tls_key_file: "key.pem"
vault:
  address: "http://localhost:8200"
  userpass_path: "userpass"
  kv_mount_path: "config/data"
  transit_path: "transit"
encrypt:
  - "regex_pattern_to_encrypt"
```

## Vault Setup:

For setup and integration with the application, follow these steps:

### 1. **Create an ACL Policy for the Project:**

Save the following policy as `project-policy.hcl`. Make sure to replace `<CONFIG: >` and `<YOUR_PROJECT_NAME>` with the correct values.

```hcl
# ==================
# KVv2 Store Paths
# ==================

# Allow full CRUD operations on the specified project within KVv2 store.
path "<CONFIG: vault.kv_mount_path>/<YOUR_PROJECT_NAME>" {
  capabilities = ["create", "update", "read", "delete", "list"]
}

# ==================
# Transit Paths
# ==================

# Allow encryption operations on the specific project encryption key.
path "<CONFIG: vault.transit_path>/encrypt/<YOUR_PROJECT_NAME>" {
  capabilities = ["create", "read", "update"]
}

# Allow decryption operations on the specific project encryption key.
path "<CONFIG: vault.transit_path>/decrypt/<YOUR_PROJECT_NAME>" {
  capabilities = ["create", "read", "update"]
}

# Allow reading encryption keys (no modification allowed).
path "<CONFIG: vault.transit_path>/keys/<YOUR_PROJECT_NAME>" {
  capabilities = ["read"]
}

# ==================
# Token Management
# ==================

# Allow tokens to manage and lookup their own properties.
path "auth/token/lookup-self" {
  capabilities = ["read"]
}
path "auth/token/renew-self" {
  capabilities = ["update"]
}
path "auth/token/revoke-self" {
  capabilities = ["update"]
}
path "sys/capabilities-self" {
  capabilities = ["update"]
}

# ==================
# Identity & Entity
# ==================

# Allow tokens to look up their own identities.
path "identity/entity/id/{{identity.entity.id}}" {
  capabilities = ["read"]
}
path "identity/entity/name/{{identity.entity.name}}" {
  capabilities = ["read"]
}

# ==================
# Response Wrapping 
# ==================

path "sys/wrapping/wrap" {
  capabilities = ["update"]
}
path "sys/wrapping/lookup" {
  capabilities = ["update"]
}
path "sys/wrapping/unwrap" {
  capabilities = ["update"]
}

```

Load the policy into Vault:
```sh
vault policy write <YOUR_PROJECT_NAME> project-policy.hcl
```

### 3. **Enable and Setup KVv2 Secret Store:**
The application retrieves its Elasticsearch configuration from the Vault KV2 engine per project basis. Enable the `kv-v2` secret engine at the desired mount path:

```sh
vault secrets enable -version=2 -path=<CONFIG: vault.kv_mount_path> kv
```

### 4. Configure Elasticsearch as KVv2 secret

 Ensure you've populated the configuration values as secrets in Vault:

```bash
vault kv put <CONFIG: vault.kv_mount_path>/<YOUR_PROJECT_NAME> \
    addresses='["https://localhost:9200"]' \
    username="<elastic-username>" \
    password="<elastic-password>" \
    state_index="<terraform-state-index>" \
    resource_index="<terraform-resources-index>" \
    lock_index="<terraform-locks-index>" \
    cloud_id="<your-cloud-id>" \
    service_token="<your-service-token>" \
    api_key="<your-api-key>" \
    certificate_fingerprint="<your-cert-fingerprint>"
```

Replace the placeholder values, such as `<elastic-username>`, with your actual configuration details.

**Note**: If you don't provide a value for any of the above fields in the command, the application will resort to using the default values.

The variables, and their default values are the following (please also refer to the Elastic cluster connection options):

1. **addresses**:
    - Type: List of strings
    - Description: Specifies the Elasticsearch cluster addresses.
    - Default: `["https://localhost:9200"]`

2. **username**:
    - Type: String
    - Description: Username for Elasticsearch authentication.
    - Default: `elastic`

3. **password**:
    - Type: String
    - Description: Password for Elasticsearch authentication.
    - Default: `elastic`

4. **state_index**:
    - Type: String
    - Description: The index name where Terraform states are stored.
    - Default: `terraform-state`

5. **resource_index**:
    - Type: String
    - Description: The index name where Terraform resources are stored.
    - Default: `terraform-resources`

6. **lock_index**:
    - Type: String
    - Description: The index name where Terraform locks are stored.
    - Default: `terraform-locks`

7. **cloud_id**:
    - Type: String
    - Description: The identifier for Elastic Cloud deployments. Use this if you're leveraging Elastic Cloud.

8. **service_token**:
    - Type: String
    - Description: An Elasticsearch service token. Use this for additional security in your Elasticsearch deployments.

9. **api_key**:
    - Type: String
    - Description: The Elasticsearch access key. Use this for authenticating to your Elasticsearch cluster.

10. **certificate_fingerprint**:
    - Type: String
    - Description: Represents the fingerprint for the Elasticsearch certificate.


### 5. **Enable and Setup Transit for Encryption:**

Enable the `transit` secret engine at the desired path:
```sh
vault secrets enable -path=<CONFIG: vault.transit_path> transit
```

Create an encryption key for your project:
```sh
vault write -f <CONFIG: vault.transit_path>/keys/<YOUR_PROJECT_NAME>
```

### 5. **Enable and Setup UserPass Authentication:**

Enable the `userpass` authentication method:
   ```sh
   vault auth enable userpass
   ```

Create a user for the application:
   ```sh
   vault write auth/userpass/users/USERNAME password=PASSWORD policies=YOUR_POLICY_NAME
   ```

## Setting up Terraform with Vault and Elasticsearch

### 1. Configure Terraform Backend for Elasticsearch:

Your application uses Elasticsearch as a backend for Terraform, so you need to configure it.

```hcl
terraform {
  backend "http" {
    address = "http://your-application-address:port/state/{project}"
    lock_address = "http://your-application-address:port/state/{project}/lock"
    unlock_address = "http://your-application-address:port/state/{project}/unlock"
    username = "${var.backend_username}"
    password = "${var.backend_password}"
  }
}
```

### 2. Initialize and Apply:

```bash
terraform init
terraform apply
```

### Contribute:

Feel free to fork this project, submit PRs, and report or fix any issues. All contributions are appreciated!

### License:
This project is under the Apache License. See [LICENSE](LICENSE) file for more details.