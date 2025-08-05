package keyring

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keyring "github.com/zalando/go-keyring"
)

func TestMain(m *testing.M) {
	keyring.MockInit()
	// Mock a keyring entry for testing
	_ = keyring.Set(Service, "foo", "bar")

	// Setup code if needed
	code := m.Run()
	// Teardown code if needed
	os.Exit(code)
}

func TestProvider_Name(t *testing.T) {
	p := &Provider{}
	assert.Equal(t, "keychain", p.Name())
}

func TestProvider_Description(t *testing.T) {
	p := &Provider{}
	want := "Use your systems keychain/keyring for storing your kubernetes and application secrets"
	assert.Equal(t, want, p.Description())
}

func TestProvider_Aliases(t *testing.T) {
	p := &Provider{}
	want := []string{"keyring", "kr"}
	got := p.Aliases()
	assert.Equal(t, want, got)
}

func TestProvider_Get_Success(t *testing.T) {
	p := &Provider{}
	val, err := p.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", val)
}

func TestProvider_Get_Error(t *testing.T) {
	p := &Provider{}
	val, err := p.Get("foobar")
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestProvider_Set_Success(t *testing.T) {
	p := &Provider{}
	err := p.Set("foobar", "barfuzz")
	assert.NoError(t, err)
}

func TestProvider_Set_Error(t *testing.T) {
	keyring.MockInitWithError(fmt.Errorf("mock error"))
	t.Cleanup(func() {
		keyring.MockInit() // Reset mock state
	})

	p := &Provider{}
	err := p.Set("foo", "baryou")
	assert.Error(t, err)
}
