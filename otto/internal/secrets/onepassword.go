// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/1password/onepassword-sdk-go"
)

// OnePasswordManager implements the Manager interface using 1Password Connect.
type OnePasswordManager struct {
	client           *onepassword.Client
	webhookSecretRef string
	appIDRef         string
	installIDRef     string
	privateKeyRef    string
	refs             map[string]string
	cachedValues     map[string]string
}

// NewOnePasswordManager creates a new OnePasswordManager with the given references.
// References should be in the format "op://vault-uuid/item-id-or-title/field".
func NewOnePasswordManager(webhookRef, appIDRef, installIDRef, privateKeyRef string) (*OnePasswordManager, error) {
	// Get token from environment variables
	token := os.Getenv("OTTO_1PASSWORD_TOKEN")
	if token == "" {
		return nil, errors.New("OTTO_1PASSWORD_TOKEN environment variable is required")
	}

	// Create the client
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Otto Bot", "v1.0.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create 1Password client: %w", err)
	}

	// Create and return the manager
	manager := &OnePasswordManager{
		client:           client,
		webhookSecretRef: webhookRef,
		appIDRef:         appIDRef,
		installIDRef:     installIDRef,
		privateKeyRef:    privateKeyRef,
		refs:             make(map[string]string),
		cachedValues:     make(map[string]string),
	}

	// Validate references
	if err := manager.validateReferences(); err != nil {
		return nil, err
	}

	return manager, nil
}

// validateReferences checks that the references are valid.
func (o *OnePasswordManager) validateReferences() error {
	if o.webhookSecretRef == "" {
		return errors.New("webhook secret reference is required")
	}

	// Either all GitHub App references must be present, or none
	hasAppID := o.appIDRef != ""
	hasInstallID := o.installIDRef != ""
	hasPrivateKey := o.privateKeyRef != ""

	if (hasAppID || hasInstallID || hasPrivateKey) &&
		(!hasAppID || !hasInstallID || !hasPrivateKey) {
		return errors.New(
			"github_app_id_ref, github_installation_id_ref, and github_private_key_ref must all be set for GitHub App authentication",
		)
	}

	return nil
}

// resolveReference gets a secret value from 1Password using the op reference.
func (o *OnePasswordManager) resolveReference(ctx context.Context, ref string) (string, error) {
	// Check cache first
	if val, ok := o.cachedValues[ref]; ok {
		return val, nil
	}

	// Resolve the reference
	value, err := o.client.Secrets().Resolve(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("failed to resolve reference %s: %w", ref, err)
	}

	// Cache the value
	o.cachedValues[ref] = value

	return value, nil
}

// GetWebhookSecret returns the GitHub webhook secret.
func (o *OnePasswordManager) GetWebhookSecret() string {
	// Check environment variable first
	if envVal := os.Getenv("OTTO_WEBHOOK_SECRET"); envVal != "" {
		return envVal
	}

	// Get the webhook secret from 1Password
	if o.webhookSecretRef != "" {
		val, err := o.resolveReference(context.Background(), o.webhookSecretRef)
		if err != nil {
			slog.Error("Failed to retrieve webhook secret from 1Password", "error", err)
			return ""
		}
		return val
	}

	return ""
}

// GetGitHubAppID returns the GitHub App ID.
func (o *OnePasswordManager) GetGitHubAppID() int64 {
	// Check environment variable first
	if envVal := os.Getenv("OTTO_GITHUB_APP_ID"); envVal != "" {
		var id int64
		_, err := fmt.Sscanf(envVal, "%d", &id)
		if err == nil && id > 0 {
			return id
		}
	}

	// Get the app ID from 1Password
	if o.appIDRef != "" {
		val, err := o.resolveReference(context.Background(), o.appIDRef)
		if err != nil {
			slog.Error("Failed to retrieve GitHub App ID from 1Password", "error", err)
			return 0
		}

		// Parse the ID
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			slog.Error("Failed to parse GitHub App ID", "error", err)
			return 0
		}

		return id
	}

	return 0
}

// GetGitHubInstallationID returns the GitHub App Installation ID.
func (o *OnePasswordManager) GetGitHubInstallationID() int64 {
	// Check environment variable first
	if envVal := os.Getenv("OTTO_GITHUB_INSTALLATION_ID"); envVal != "" {
		var id int64
		_, err := fmt.Sscanf(envVal, "%d", &id)
		if err == nil && id > 0 {
			return id
		}
	}

	// Get the installation ID from 1Password
	if o.installIDRef != "" {
		val, err := o.resolveReference(context.Background(), o.installIDRef)
		if err != nil {
			slog.Error("Failed to retrieve GitHub Installation ID from 1Password", "error", err)
			return 0
		}

		// Parse the ID
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			slog.Error("Failed to parse GitHub Installation ID", "error", err)
			return 0
		}

		return id
	}

	return 0
}

// GetGitHubPrivateKey returns the GitHub App private key.
func (o *OnePasswordManager) GetGitHubPrivateKey() []byte {
	// Check environment variable first
	if envVal := os.Getenv("OTTO_GITHUB_PRIVATE_KEY"); envVal != "" {
		return []byte(envVal)
	}

	// Get the private key from 1Password
	if o.privateKeyRef != "" {
		val, err := o.resolveReference(context.Background(), o.privateKeyRef)
		if err != nil {
			slog.Error("Failed to retrieve GitHub private key from 1Password", "error", err)
			return nil
		}

		return []byte(val)
	}

	return nil
}

// LoadOnePasswordConfig loads 1Password configuration from the given path.
func LoadOnePasswordConfig(path string) (*OnePasswordManager, error) {
	// Read the configuration file
	slog.Info("Loading 1Password configuration from file", "path", path)

	// Parse op:// references
	// For now, this is a placeholder. In a complete implementation,
	// this would parse a YAML config file with the op:// references.
	return nil, errors.New("not implemented")
}
