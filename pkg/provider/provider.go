// Package provider defines the interface that must be implemented by anypassword manager provider.
// It specifies methods for retrieving and storing credentials, as well as obtaining metadata
// such as the provider's name, description, and alternative aliases.
package provider

// Provider represents a password manager provider interface.
type Provider interface {
	// Get retrieves a credential from the provider
	Get(itemName string) (string, error)
	// Set stores a credential in the provider
	Set(itemName, secret string) error
	// Name returns the name of the provider
	Name() string
	// Description returns a description of the provider
	Description() string
	// Aliases returns alternative names for the provider
	Aliases() []string
}
