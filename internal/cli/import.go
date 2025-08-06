// Package cli provides import functionality for migrating kubeconfig credentials
// to password managers. This module handles backing up existing kubeconfigs,
// extracting credentials, storing them in the selected provider, and generating
// new kubeconfig files with exec plugin configurations.
package cli

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/passman"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Create a local alias for Kubeconfig.
type KubeConfig = api.Config

// Create a local alias for UserInfo.
type UserInfo = api.AuthInfo

var (
	// ErrCredentialAlreadyExists indicates a credential already exists in the provider.
	ErrCredentialAlreadyExists = errors.New("credential already exists")
	// ErrNoCredentialsFound indicates no credentials were found for a user.
	ErrNoCredentialsFound = errors.New("no credentials found for user")
	// ErrUserNotFound indicates the specified user was not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrCredentialValidation indicates credential validation failed.
	ErrCredentialValidation = errors.New("credential validation failed")
)

// ImportCommand creates the import command for migrating kubeconfig credentials.
func ImportCommand() cli.Command {
	return cli.Command{
		Name:  "import",
		Usage: "Import kubeconfig credentials into a password manager",
		Description: `This command will:
1. Backup your existing kubeconfig file
2. Extract credentials from users with direct auth (token, certificates)
3. Store them in the specified password manager
4. Generate a new kubeconfig with exec plugin configurations
5. Replace your kubeconfig with the new version

The original kubeconfig will be backed up with a timestamp.`,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "kubeconfig, k",
				Value: getDefaultKubeconfigPath(),
				Usage: "Path to kubeconfig file",
			},
			cli.StringFlag{
				Name:  "provider, p",
				Usage: "Password manager provider to use (required)",
			},
			cli.StringFlag{
				Name:  "prefix",
				Value: "kubectl-passman",
				Usage: "Prefix for stored credential names",
			},
			cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be done without making changes",
			},
			cli.BoolFlag{
				Name:  "force",
				Usage: "Overwrite existing credentials in the password manager",
			},
		},
		Action: handleImportCommand,
	}
}

// handleImportCommand handles the import command execution.
func handleImportCommand(ctx *cli.Context) error {
	config, err := parseImportConfig(ctx)
	if err != nil {
		return err
	}

	prov, err := validateProvider(config.providerName)
	if err != nil {
		return err
	}

	printImportHeader(config)

	kubeconfig, usersToMigrate, err := prepareImport(config.kubeconfigPath)
	if err != nil {
		return err
	}

	if len(usersToMigrate) == 0 {
		fmt.Println("No users with direct credentials found to migrate")

		return nil
	}

	printUsersToMigrate(usersToMigrate)

	fmt.Println("\n=== Starting Import Process ===")

	if err := handleBackup(config); err != nil {
		return err
	}

	fmt.Printf("\n=== Starting Migration of %d user(s) ===\n", len(usersToMigrate))

	if err := processUserMigrations(kubeconfig, usersToMigrate, config, prov); err != nil {
		return err
	}

	fmt.Println("\n=== Finalizing Import ===")

	if err := finalizeImport(config, kubeconfig); err != nil {
		return err
	}

	fmt.Println("\n✅ Import completed successfully!")

	return nil
}

// importConfig holds all configuration for the import operation.
type importConfig struct {
	providerName   string
	kubeconfigPath string
	prefix         string
	dryRun         bool
	force          bool
}

// parseImportConfig extracts and validates command line arguments.
func parseImportConfig(ctx *cli.Context) (*importConfig, error) {
	providerName := ctx.String("provider")
	if providerName == "" {
		return nil, cli.NewExitError("Provider is required. Use --provider flag", 1)
	}

	return &importConfig{
		providerName:   providerName,
		kubeconfigPath: ctx.String("kubeconfig"),
		prefix:         ctx.String("prefix"),
		dryRun:         ctx.Bool("dry-run"),
		force:          ctx.Bool("force"),
	}, nil
}

// validateProvider checks if the specified provider exists.
func validateProvider(providerName string) (provider.Provider, error) {
	prov, exists := registry.GetProvider(providerName)
	if !exists {
		availableProviders := getAvailableProviders()

		return nil, cli.NewExitError(
			fmt.Sprintf("Provider '%s' not found. Available providers: %s",
				providerName, strings.Join(availableProviders, ", ")), 1)
	}

	return prov, nil
}

// printImportHeader displays the import configuration to the user.
func printImportHeader(config *importConfig) {
	fmt.Printf("Importing kubeconfig credentials from: %s\n", config.kubeconfigPath)
	fmt.Printf("Using provider: %s\n", config.providerName)
	fmt.Printf("Credential prefix: %s\n", config.prefix)

	if config.dryRun {
		fmt.Println("DRY RUN - No changes will be made")
	}
}

// prepareImport loads kubeconfig and identifies users to migrate.
func prepareImport(kubeconfigPath string) (kubeconfig *KubeConfig, users []string, err error) {
	kubeconfig, err = clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, nil, cli.NewExitError(fmt.Sprintf("Failed to load kubeconfig: %v", err), 1)
	}

	users = findUsersWithCredentials(kubeconfig)

	return kubeconfig, users, nil
}

// printUsersToMigrate displays the list of users that will be migrated.
func printUsersToMigrate(usersToMigrate []string) {
	fmt.Printf("Found %d user(s) with credentials to migrate:\n", len(usersToMigrate))

	for _, userName := range usersToMigrate {
		fmt.Printf("  - %s\n", userName)
	}
}

// handleBackup creates a backup of the original kubeconfig if not in dry-run mode.
func handleBackup(config *importConfig) error {
	if config.dryRun {
		fmt.Println("⏭️  Skipping backup (dry-run mode)")
		return nil
	}

	fmt.Println("💾 Creating backup of original kubeconfig...")

	backupPath, err := backupKubeconfig(config.kubeconfigPath)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to backup kubeconfig: %v", err), 1)
	}

	fmt.Printf("✅ Backed up original kubeconfig to: %s\n", backupPath)

	return nil
}

// processUserMigrations handles the migration of all users.
func processUserMigrations(
	kubeconfig *KubeConfig,
	usersToMigrate []string,
	config *importConfig,
	prov provider.Provider,
) error {
	migratedCount := 0
	skippedCount := 0

	for i, userName := range usersToMigrate {
		credentialName := fmt.Sprintf("%s-%s", config.prefix, userName)

		fmt.Printf("🔄 [%d/%d] Migrating user '%s'...\n", i+1, len(usersToMigrate), userName)

		if config.dryRun {
			fmt.Printf("   ⏭️  Would migrate to credential '%s' (dry-run)\n", credentialName)

			continue
		}

		err := migrateUser(kubeconfig,
			userName,
			credentialName,
			prov,
			config.providerName,
			config.force,
		)
		if err != nil {
			// Check if it's just a credential already exists error
			if errors.Is(err, ErrCredentialAlreadyExists) {
				fmt.Printf("   ⚠️  Credential '%s' already exists (skipping, use --force to overwrite)\n", credentialName)
				skippedCount++
				continue
			}
			// Handle validation errors
			if errors.Is(err, ErrCredentialValidation) || errors.Is(err, ErrNoCredentialsFound) {
				fmt.Printf("   ⚠️  Failed to process credentials for '%s': %v (skipping)\n", userName, err)
				skippedCount++
				continue
			}
			return fmt.Errorf("Failed to migrate user: %w", err)
		}

		fmt.Printf("   ✅ Successfully migrated to credential '%s'\n", credentialName)
		migratedCount++
	}

	if !config.dryRun {
		if skippedCount > 0 {
			fmt.Printf("✅ Migration completed: %d migrated, %d skipped (errors or already exist)\n", migratedCount, skippedCount)
		} else {
			fmt.Printf("✅ All %d user(s) migrated successfully!\n", migratedCount)
		}
	}

	return nil
}

// finalizeImport writes the updated kubeconfig if not in dry-run mode.
func finalizeImport(config *importConfig, kubeconfig *KubeConfig) error {
	if config.dryRun {
		fmt.Println("⏭️  Skipping kubeconfig update (dry-run mode)")
		return nil
	}

	fmt.Println("📝 Writing updated kubeconfig...")

	err := writeKubeconfig(config.kubeconfigPath, kubeconfig)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to write new kubeconfig: %v", err), 1)
	}

	fmt.Printf("✅ Successfully updated kubeconfig at: %s\n", config.kubeconfigPath)

	return nil
}

// getDefaultKubeconfigPath returns the default kubeconfig path.
func getDefaultKubeconfigPath() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Clean(filepath.Join(homeDir, ".kube", "config"))
}

// loadKubeconfig loads and parses a kubeconfig file.
func loadKubeconfig(path string) (*KubeConfig, error) {
	// Validate the file path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: %s", path)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig file: %w", err)
	}

	var kubeconfig KubeConfig

	err = yaml.Unmarshal(data, &kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig YAML: %w", err)
	}

	return &kubeconfig, nil
}

// findUsersWithCredentials identifies users that have direct credentials.
func findUsersWithCredentials(kubeconfig *KubeConfig) []string {
	var users []string

	for name, user := range kubeconfig.AuthInfos {
		// We ignore existing Exec implementations
		if user.Exec != nil {
			continue
		}

		hasCredentials := user.Token != "" ||
			len(user.ClientCertificateData) != 0 ||
			user.ClientCertificate != "" ||
			len(user.ClientKeyData) != 0 ||
			user.ClientKey != "" ||
			(user.Username != "" && user.Password != "")

		// Skip users that already use exec plugins
		if hasCredentials {
			users = append(users, name)
		}
	}

	return users
}

const KubeConfigFileMode os.FileMode = 0o600 // Default file mode for kubeconfig (Backup) files

// backupKubeconfig creates a backup of the original kubeconfig file.
func backupKubeconfig(originalPath string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")

	backupPath := fmt.Sprintf("%s/kubeconfig-backup-%s", os.TempDir(), timestamp)

	data, err := os.ReadFile(filepath.Clean(originalPath))
	if err != nil {
		return "", fmt.Errorf("failed to read original kubeconfig: %w", err)
	}

	err = os.WriteFile(backupPath, data, KubeConfigFileMode)
	if err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// migrateUser migrates a single user's credentials to the password manager.
func migrateUser(
	kubeconfig *KubeConfig,
	userName, credentialName string,
	prov provider.Provider,
	providerName string,
	force bool,
) error {
	// Find the user
	var userInfo *UserInfo

	for name, info := range kubeconfig.AuthInfos {
		if name == userName {
			userInfo = info

			break
		}
	}

	if userInfo == nil {
		return fmt.Errorf("%w: %s", ErrUserNotFound, userName)
	}

	// // Extract credentials
	// credentials := extractCredentials(userInfo)
	// if len(credentials) == 0 {
	// 	return fmt.Errorf("%w: %s", ErrNoCredentialsFound, userName)
	// }

	// Convert to JSON for storage
	credentialJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Validate the credential format
	validatedCredential, err := passman.FormatValidator(string(credentialJSON))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCredentialValidation, err)
	}

	// Check if credential already exists (unless force is used)
	if !force {
		_, err := prov.Get(credentialName)
		if err == nil {
			return fmt.Errorf("%w: %s", ErrCredentialAlreadyExists, credentialName)
		}
	}

	// Store in password manager
	err = prov.Set(credentialName, validatedCredential)
	if err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	// Update user to use exec plugin
	updateUserToExecPlugin(userInfo, credentialName, providerName)

	return nil
}

// extractCredentials extracts credentials from a UserInfo struct.
func extractCredentials(userInfo *UserInfo) map[string]string {
	credentials := make(map[string]string)

	if userInfo.Token != "" {
		credentials["token"] = userInfo.Token
	}

	// Handle certificate data - these are already decoded bytes, need to base64 encode for storage
	if len(userInfo.ClientCertificateData) != 0 {
		credentials["client-certificate-data"] = base64.StdEncoding.EncodeToString(userInfo.ClientCertificateData)
	}

	if len(userInfo.ClientKeyData) != 0 {
		credentials["client-key-data"] = base64.StdEncoding.EncodeToString(userInfo.ClientKeyData)
	}

	// Handle certificate files - these are file paths
	if userInfo.ClientCertificate != "" {
		credentials["client-certificate"] = userInfo.ClientCertificate
	}

	if userInfo.ClientKey != "" {
		credentials["client-key"] = userInfo.ClientKey
	}

	if userInfo.Username != "" && userInfo.Password != "" {
		credentials["username"] = userInfo.Username
		credentials["password"] = userInfo.Password
	}

	return credentials
}

// updateUserToExecPlugin updates a user configuration to use the exec plugin.
func updateUserToExecPlugin(userInfo *UserInfo, credentialName, providerName string) {
	// Clear existing credentials
	userInfo.Token = ""
	userInfo.ClientCertificateData = []byte{}
	userInfo.ClientKeyData = []byte{}
	userInfo.ClientCertificate = ""
	userInfo.ClientKey = ""
	userInfo.Username = ""
	userInfo.Password = ""
	userInfo.AuthProvider = nil

	// Set up exec plugin
	userInfo.Exec = &api.ExecConfig{
		APIVersion:         "client.authentication.k8s.io/v1beta1",
		Command:            "kubectl-passman",
		Args:               []string{providerName, credentialName},
		ProvideClusterInfo: true,
		InteractiveMode:    api.NeverExecInteractiveMode,
	}
}

// writeKubeconfig writes a kubeconfig to a file.
func writeKubeconfig(path string, kubeconfig *api.Config) error {
	return clientcmd.WriteToFile(*kubeconfig, path)
}

// getAvailableProviders returns a list of available provider names.
func getAvailableProviders() []string {
	providers := registry.GetAllProviders()

	var names []string

	seen := make(map[string]bool)

	for _, prov := range providers {
		// Only include primary names (not aliases)
		if !seen[prov.Name()] {
			names = append(names, prov.Name())
			seen[prov.Name()] = true
		}
	}

	return names
}
