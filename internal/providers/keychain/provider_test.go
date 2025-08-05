package keychain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseProvider_Name(t *testing.T) {
	p := &Base{}
	if got := p.Name(); got != "keychain" {
		t.Errorf("Name() = %q, want %q", got, "keychain")
	}
}

func TestBaseProvider_Description(t *testing.T) {
	p := &Base{}
	want := "Use your systems keychain/keyring for storing your kubernetes and application secrets"
	if got := p.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestBaseProvider_Aliases(t *testing.T) {
	p := &Base{}
	want := []string{"keyring"}
	got := p.Aliases()
	assert.Len(t, got, len(want))
	assert.Equal(t, want, got)
}

type mockKeychain struct {
	getCalled bool
	setCalled bool
	getVal    string
	getErr    error
	setErr    error
}

func (m *mockKeychain) Get(itemName string) (string, error) {
	m.getCalled = true
	return m.getVal, m.getErr
}

func (m *mockKeychain) Set(itemName, secret string) error {
	m.setCalled = true
	return m.setErr
}

func TestBaseProvider_Get_WithImpl(t *testing.T) {
	p := &mockKeychain{getVal: "secret", getErr: nil}
	val, err := p.Get("foo")
	require.NoError(t, err, "Get() should not return an error")

	assert.Equal(t, "secret", val, "Get() should return the expected value")
	assert.True(t, p.getCalled, "Get() should call underlying impl")
}

func TestBaseProvider_Get_NoImpl(t *testing.T) {
	p := &Base{}
	val, err := p.Get("foo")

	assert.Empty(t, val, "Get() should return NotImplemented error when impl is nil")
	assert.Error(t, err)
}

func TestBaseProvider_Set_WithImpl(t *testing.T) {
	p := &mockKeychain{setErr: ErrNotImplemented}
	err := p.Set("foo", "bar")

	require.Error(t, err)
	assert.True(t, p.setCalled, "Set() should call underlying impl")
}

func TestBaseProvider_Set_NoImpl(t *testing.T) {
	p := &Base{}
	err := p.Set("foo", "bar")

	require.Error(t, err, "Set() should return an error when impl is nil")
	assert.Equal(t, err, ErrNotImplemented)
}
