package onepassword

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_Name(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "1password", provider.Name())
}

func TestProvider_Description(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "Use 1Password Connect for storing your kubernetes authentication secrets", provider.Description())
}

func TestProvider_Aliases(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, []string{"1pass", "op"}, provider.Aliases())
}

func TestProvider_Get_NoEnvironmentVariables(t *testing.T) {
	provider := &Provider{}

	// Without environment variables, it should fail
	_, err := provider.Get("test-item")
	require.Error(t, err)
	require.Contains(t, err.Error(), "environment variables not set")
}

func TestProvider_Set_NoEnvironmentVariables(t *testing.T) {
	provider := &Provider{}

	// Without environment variables, it should fail
	err := provider.Set("test-item", "test-secret")
	require.Error(t, err)
	require.Contains(t, err.Error(), "environment variables not set")
}
