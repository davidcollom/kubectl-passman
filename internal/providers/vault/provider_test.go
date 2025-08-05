package vault

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_Name(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "vault", provider.Name())
}

func TestProvider_Description(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "Use HashiCorp Vault for storing your kubernetes and application secrets", provider.Description())
}

func TestProvider_Aliases(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, []string{"hashicorp-vault", "hcv"}, provider.Aliases())
}

func TestProvider_Get_NoEnvironmentVariables(t *testing.T) {
	provider := &Provider{}

	// Without environment variables, it should fail
	_, err := provider.Get("secret/test:password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "vault environment variables not set")
}

func TestProvider_Set_NoEnvironmentVariables(t *testing.T) {
	provider := &Provider{}

	// Without environment variables, it should fail
	err := provider.Set("secret/test:password", "test-secret")
	require.Error(t, err)
	require.Contains(t, err.Error(), "vault environment variables not set")
}

func TestProvider_ParseSecretPath(t *testing.T) {
	provider := &Provider{}

	tests := []struct {
		name     string
		input    string
		wantPath string
		wantKey  string
	}{
		{
			name:     "path with key",
			input:    "secret/data/myapp:password",
			wantPath: "secret/data/myapp",
			wantKey:  "password",
		},
		{
			name:     "path without key (default)",
			input:    "secret/data/myapp",
			wantPath: "secret/data/myapp",
			wantKey:  "password",
		},
		{
			name:     "path with custom key",
			input:    "secret/data/myapp:token",
			wantPath: "secret/data/myapp",
			wantKey:  "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotKey := provider.parseSecretPath(tt.input)
			require.Equal(t, tt.wantPath, gotPath)
			require.Equal(t, tt.wantKey, gotKey)
		})
	}
}
