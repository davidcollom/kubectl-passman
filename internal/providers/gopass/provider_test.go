package gopass

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGopassProvider_Name(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "gopass", provider.Name())
}

func TestGopassProvider_Description(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, "Use gopass for storing your kubernetes and application secrets", provider.Description())
}

func TestGopassProvider_Aliases(t *testing.T) {
	provider := &Provider{}
	require.Equal(t, []string{}, provider.Aliases())
}

func TestGopassProvider_Get(t *testing.T) {
	// Note: This test would require mocking the exec.Command call
	// For now, we just test that the interface is implemented correctly
	provider := &Provider{}
	require.NotNil(t, provider)

	// We can't easily test the actual functionality without mocking
	// the command execution, but we can verify the method exists
	_, err := provider.Get("test-item")
	// This will likely fail since gopass isn't installed, but that's expected
	require.Error(t, err)
}

func TestGopassProvider_Set(t *testing.T) {
	provider := &Provider{}
	require.NotNil(t, provider)

	// Similarly, this will likely fail without gopass installed
	err := provider.Set("test-item", "test-secret")
	require.Error(t, err)
}
