# Otto, a helpful bot

![otto, a helpful otter, is here to help you!](./bin/otto.png)

Otto is a Golang-based GitHub bot that's built to assist OpenTelemetry maintainers with various tasks.

## What Can Otto Do?

Otto provides a variety of features. Features are provided by modules.

Right now, the only feature is 'oncall', which helps manage on-call rotations for repository maintainers.

- **oncall**: Assigns tasks to on-call users, tracks acknowledgment, and handles escalations

## Installation

### Prerequisites

- Go 1.24+
- SQLite
- GitHub App (for authentication)

### Configuration

Otto supports several methods for configuration, with a focus on securely managing sensitive information.

#### Application Configuration

**config.yaml**: Non-sensitive application configuration
- Server port, database path, logging settings, module configuration
- See `config.example.yaml` for an example

#### Secrets Configuration

Otto supports three methods for managing secrets, in order of preference:

1. **1Password Integration**
   - Securely stores secrets in a 1Password vault
   - Requires 1Password account and Connect API:
     - 1Password Connect Server (self-hosted or cloud)
     - 1Password Connect API token
   - Configure via `onepassword.example.yaml` and set `OTTO_1PASSWORD_CONFIG` env var
   - Required env vars: `OTTO_1PASSWORD_URL`, `OTTO_1PASSWORD_TOKEN`

2. **Secrets File**
   - YAML file containing sensitive information (webhook secret, GitHub credentials)
   - See `secrets.example.yaml` for an example
   - Use this for development or simple deployments

3. **Environment Variables**
   - Fallback method that works with any deployment
   - Available variables:
     - `OTTO_WEBHOOK_SECRET`: GitHub webhook secret
     - `OTTO_GITHUB_APP_ID`: GitHub App ID
     - `OTTO_GITHUB_INSTALLATION_ID`: GitHub App Installation ID
     - `OTTO_GITHUB_PRIVATE_KEY`: GitHub App private key (the actual key content)

### GitHub App Setup

1. Create a GitHub App at `https://github.com/settings/apps/new`
2. Configure the permissions:
   - Repository permissions: 
     - Issues: Read & Write
     - Pull requests: Read & Write
     - Metadata: Read-only
   - Subscribe to events:
     - Issues
     - Issue comments
     - Pull requests
3. Generate a private key and download it
4. Install the app on your repositories
5. Note the App ID and Installation ID
6. Configure Otto with these values

### Running Otto

```bash
# Build the application
go build -o otto ./cmd/otto

# Run with default config paths (config.yaml, secrets.yaml)
./otto

# Run with custom config paths
OTTO_CONFIG=custom-config.yaml OTTO_SECRETS=custom-secrets.yaml ./otto
```

### Health Checks

Otto provides the following HTTP endpoints for health monitoring:

- `/check/liveness` - Kubernetes liveness probe (checks if the server can process requests)
- `/check/readiness` - Kubernetes readiness probe (checks if all dependencies are ready, including database connectivity)

Use these endpoints for monitoring and orchestration platforms:

```bash
# Kubernetes example:
livenessProbe:
  httpGet:
    path: /check/liveness
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /check/readiness
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

### Docker

You can run Otto using Docker with any of the supported configuration methods:

#### Using environment variables (recommended for production)

```bash
docker run -p 8080:8080 \
  -v /path/to/config.yaml:/home/otto/config.yaml \
  -e OTTO_WEBHOOK_SECRET="your_webhook_secret" \
  -e OTTO_GITHUB_APP_ID="123456" \
  -e OTTO_GITHUB_INSTALLATION_ID="789012" \
  -e OTTO_GITHUB_PRIVATE_KEY="$(cat /path/to/private-key.pem)" \
  otto:latest
```

#### Using secrets file

```bash
docker run -p 8080:8080 \
  -v /path/to/config.yaml:/home/otto/config.yaml \
  -v /path/to/secrets.yaml:/home/otto/secrets.yaml \
  otto:latest
```

#### Using 1Password Connect (most secure)

```bash
docker run -p 8080:8080 \
  -v /path/to/config.yaml:/home/otto/config.yaml \
  -v /path/to/onepassword.yaml:/home/otto/onepassword.yaml \
  -e OTTO_1PASSWORD_CONFIG="/home/otto/onepassword.yaml" \
  -e OTTO_1PASSWORD_URL="https://your-1password-connect-server" \
  -e OTTO_1PASSWORD_TOKEN="your_1password_connect_token" \
  otto:latest
```

This requires:
1. A 1Password Connect server (either self-hosted or using 1Password Cloud)
2. A 1Password Connect API token
3. Items in your 1Password vault for each secret (webhook secret, GitHub App credentials)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to Otto.
