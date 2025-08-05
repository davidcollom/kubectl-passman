package passman

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	jsonCertBase64 = `{"client-certificate-data":"MDAwMDA=","client-key-data":"MDAwMDA="}`
	jsonCert       = `{"clientCertificateData":"00000","clientKeyData":"00000"}`
	jsonToken      = `{"token":"00000"}`
)

func TestFormatValidatorCertBase64(t *testing.T) {
	actual, err := FormatValidator(jsonCertBase64)
	require.Equal(t, jsonCert, actual)
	require.Nil(t, err)
}

func TestFormatValidatorCertBase64ErrorDecode(t *testing.T) {
	actual, err := FormatValidator(`{"client-certificate-data":"BAD-DATA","client-key-data":"MDAwMDA="}`)
	require.Equal(t, "", actual)
	require.Equal(t, "illegal base64 data at input byte 3", err.Error())
}

func TestFormatValidatorCertBase64ErrorMisKey(t *testing.T) {
	actual, err := FormatValidator(`{"clientCertificateData":"BAD-DATA","client-certificate-data":"MDAwMDA="}`)
	require.Equal(t, "", actual)
	require.Equal(t, "cannot define valid secret format", err.Error())
}

func TestFormatValidatorCert(t *testing.T) {
	actual, err := FormatValidator(jsonCert)
	require.Equal(t, jsonCert, actual)
	require.Nil(t, err)
}

func TestFormatValidatorToken(t *testing.T) {
	actual, err := FormatValidator(jsonToken)
	require.Equal(t, jsonToken, actual)
	require.Nil(t, err)
}

func TestFormatValidatorEmpty(t *testing.T) {
	actual, err := FormatValidator(`{}`)
	require.Equal(t, "", actual)
	require.Equal(t, "cannot define valid secret format", err.Error())
}

func TestFormatValidatorBadJSON(t *testing.T) {
	actual, err := FormatValidator(`{bad json}`)
	require.Equal(t, "", actual)
	require.NotNil(t, err)
}

func TestFormatResponse(t *testing.T) {
	res := &Response{
		Status: ResponseStatus{
			Token: "test-token",
		},
	}
	actual, err := FormatResponse(res)
	require.Nil(t, err)
	expected := `{"apiVersion":"client.authentication.k8s.io/v1beta1","kind":"ExecCredential","status":{"token":"test-token"}}`
	require.Equal(t, expected, actual)
}
