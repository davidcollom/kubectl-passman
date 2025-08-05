//go:build !darwin && !linux && !windows
// +build !darwin,!linux,!windows

package keychain

import "github.com/chrisns/kubectl-passman/internal/registry"

// Provider provides a fallback implementation for unsupported platforms
type Provider struct {
	Base // Embed the base provider for metadata methods
}

func init() {
	registry.Register(&Provider{})
}

// Get returns an error for unsupported platforms (Base.Get already handles this)
// Set returns an error for unsupported platforms (Base.Set already handles this)
