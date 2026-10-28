package keyring

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keyring "github.com/zalando/go-keyring"
)

var p = &Provider{} //nolint: varnamelen

var ErrMockError = errors.New("mock error")

func init() {
	// Ensure we Init Keyring with a Mock....
	keyring.MockInit()
}

func TestMain(m *testing.M) {
	// Mock a keyring entry for testing
	err := keyring.Set(serviceName, "foo", "bar")
	if err != nil {
		panic(fmt.Sprintf("Failed to set mock keyring entry: %v\n", err))
	}

	// Setup code if needed
	code := m.Run()
	// Teardown code if needed
	os.Exit(code)
}

func TestProvider_Metadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(*testing.T, *Provider)
	}{
		{
			name: "Name returns keychain",
			testFunc: func(t *testing.T, p *Provider) {
				t.Helper()

				assert.Equal(t, "keychain", p.Name())
			},
		},
		{
			name: "Description returns correct value",
			testFunc: func(t *testing.T, p *Provider) {
				t.Helper()

				want := "Use your systems keychain/keyring for storing your kubernetes and application secrets"
				assert.Equal(t, want, p.Description())
			},
		},
		{
			name: "Aliases returns correct values",
			testFunc: func(t *testing.T, p *Provider) {
				t.Helper()

				want := []string{"keyring", "kr"}
				got := p.Aliases()
				assert.Equal(t, want, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.testFunc(t, p)
		})
	}
}

func TestProvider_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantError bool
	}{
		{
			name:      "Success - retrieves existing key",
			key:       "foo",
			wantValue: "bar",
			wantError: false,
		},
		{
			name: "Error - key does not exist",
			key: strings.Join(
				[]string{t.Name(), "nonexistent", strconv.FormatInt(time.Now().Unix(), 10)},
				"-",
			),
			wantValue: "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Ensure _Something_ is in the keychain!
			if !tt.wantError {
				require.NoError(t, p.Set(tt.key, tt.wantValue))
			}

			val, err := p.Get(tt.key)

			if tt.wantError {
				require.Error(t, err)
				assert.Empty(t, val)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, val)
			}
		})
	}
}

func TestProvider_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       string
		value     string
		setupFunc func(*testing.T)
		wantError bool
		checkErr  func(*testing.T, error)
	}{
		{
			name:      "Success - sets value",
			key:       "foobar",
			value:     "barfuzz",
			wantError: false,
		},
		{
			name:  "Error - mock returns error",
			key:   "foo",
			value: "baryou",
			setupFunc: func(t *testing.T) {
				t.Helper()

				keyring.MockInitWithError(ErrMockError)

				t.Cleanup(func() {
					keyring.MockInit() // Reset mock state
				})
			},
			wantError: true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()

				assert.Equal(t, ErrMockError, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			err := p.Set(tt.key, tt.value)

			if tt.wantError {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
