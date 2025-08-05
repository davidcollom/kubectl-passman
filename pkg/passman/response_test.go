package passman

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	jsonCertBase64 = `{"client-certificate-data":"MDAwMDA=","client-key-data":"MDAwMDA="}`
	jsonCert       = `{"clientCertificateData":"00000","clientKeyData":"00000"}`
	jsonToken      = `{"token":"00000"}`
)

func TestFormatValidatorCertBase64(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(jsonCertBase64)
	require.JSONEq(t, jsonCert, actual)
	require.NoError(t, err)
}

func TestFormatValidatorCertBase64ErrorDecode(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(
		`{"client-certificate-data":"BAD-DATA","client-key-data":"MDAwMDA="}`,
	)
	require.Empty(t, actual)
	require.Contains(t, err.Error(), "illegal base64 data at input byte")
}

func TestFormatValidatorCertBase64ErrorMisKey(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(
		`{"clientCertificateData":"BAD-DATA","client-certificate-data":"MDAwMDA="}`,
	)
	require.Empty(t, actual)
	require.Contains(t, err.Error(), "cannot define valid secret format")
}

func TestFormatValidatorCert(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(jsonCert)
	require.JSONEq(t, jsonCert, actual)
	require.NoError(t, err)
}

func TestFormatValidatorToken(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(jsonToken)
	require.JSONEq(t, jsonToken, actual)
	require.NoError(t, err)
}

func TestFormatValidatorEmpty(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(`{}`)
	require.Empty(t, actual)
	require.Contains(t, err.Error(), "cannot define valid secret format")
}

func TestFormatValidatorBadJSON(t *testing.T) {
	t.Parallel()

	actual, err := FormatValidator(`{bad json}`)
	require.Empty(t, actual)
	require.Error(t, err)
}

func TestFormatResponse(t *testing.T) {
	t.Parallel()

	res := &Response{
		Status: ResponseStatus{
			Token: "test-token",
		},
	}
	actual, err := FormatResponse(res)
	require.NoError(t, err)

	expected := strings.TrimSpace(`
		{"apiVersion":"client.authentication.k8s.io/v1beta1","kind":"ExecCredential","status":{"token":"test-token"}}
	`)
	require.Equal(t, expected, actual)
}
