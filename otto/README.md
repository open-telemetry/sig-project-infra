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

Otto uses two separate configuration files:

1. **config.yaml**: Non-sensitive application configuration
   - Server port, database path, logging settings, module configuration
   - See `config.example.yaml` for an example

2. **secrets.yaml**: Sensitive authentication information
   - GitHub webhook secret
   - GitHub App credentials (app ID, installation ID, private key)
   - See `secrets.example.yaml` for an example

You can also provide sensitive configuration via environment variables:

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

### Docker

You can also run Otto using Docker:

```bash
# Build the Docker image
docker build -t otto:latest .

# Run the container
docker run -p 8080:8080 \
  -v /path/to/config.yaml:/home/otto/config.yaml \
  -v /path/to/secrets.yaml:/home/otto/secrets.yaml \
  otto:latest
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to Otto.
