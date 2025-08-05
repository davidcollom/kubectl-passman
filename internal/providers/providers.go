// Package providers imports all available provider implementations to ensure their
// init() functions are executed, registering each provider with the central registry.
// This design allows the CLI to access all supported providers through a single import.
package providers

import (
	_ "github.com/chrisns/kubectl-passman/internal/providers/conjur"
	_ "github.com/chrisns/kubectl-passman/internal/providers/gopass"
	_ "github.com/chrisns/kubectl-passman/internal/providers/keyring"
	_ "github.com/chrisns/kubectl-passman/internal/providers/onepassword"
	_ "github.com/chrisns/kubectl-passman/internal/providers/vault"
)
