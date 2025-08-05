package providers

// This package imports all provider implementations to trigger their init() functions
// and register them with the registry. This provides a single import point for the CLI.

import (
	_ "github.com/chrisns/kubectl-passman/internal/providers/conjur"
	_ "github.com/chrisns/kubectl-passman/internal/providers/gopass"
	_ "github.com/chrisns/kubectl-passman/internal/providers/keychain"
	_ "github.com/chrisns/kubectl-passman/internal/providers/onepassword"
	_ "github.com/chrisns/kubectl-passman/internal/providers/vault"
)
