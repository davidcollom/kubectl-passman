package conjur

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_Name(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "conjur", provider.Name())
}

func TestProvider_Description(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "Use CyberArk Conjur for storing your kubernetes and application secrets", provider.Description())
}

func TestProvider_Aliases(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, []string{"ca"}, provider.Aliases())
}

func TestProvider_Get_NoConfiguration(t *testing.T) {
	provider := &Provider{}

	// Without configuration, it should fail
	_, err := provider.Get("test-variable")
	require.Error(t, err)
	require.Contains(t, err.Error(), "conjur environment variables not set")
}

func TestProvider_Set_NotSupported(t *testing.T) {
	provider := &Provider{}

	// Set operation is not supported for Conjur
	err := provider.Set("test-variable", "test-value")
	require.Error(t, err)
	require.Contains(t, err.Error(), "conjur environment variables not set")
}
