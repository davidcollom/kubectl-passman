package passman

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/creasty/defaults"
)

// ResponseStatus represents the status field of an ExecCredential response
type ResponseStatus struct {
	Token                  string `json:"token,omitempty"`
	ClientCertificateData  string `json:"clientCertificateData,omitempty"`
	ClientCertificateDataD string `json:"client-certificate-data,omitempty"`
	ClientKeyData          string `json:"clientKeyData,omitempty"`
	ClientKeyDataD         string `json:"client-key-data,omitempty"`
}

// Response represents a Kubernetes ExecCredential response
type Response struct {
	APIVersion string         `default:"client.authentication.k8s.io/v1beta1" json:"apiVersion"`
	Kind       string         `default:"ExecCredential" json:"kind"`
	Status     ResponseStatus `json:"status"`
}

// FormatValidator validates and normalizes the secret format
func FormatValidator(secret string) (string, error) {
	s := &ResponseStatus{}
	data := []byte(secret)

	err := json.Unmarshal(data, s)
	if err != nil {
		return "", err
	}

	switch {
	case len(s.ClientCertificateDataD) > 0 && len(s.ClientKeyDataD) > 0:
		dataCrt, errCrt := base64.StdEncoding.DecodeString(s.ClientCertificateDataD)
		dataKey, errKey := base64.StdEncoding.DecodeString(s.ClientKeyDataD)

		switch {
		case errCrt != nil:
			return "", errCrt
		case errKey != nil:
			return "", errKey
		default:
			s.ClientCertificateData = string(dataCrt)
			s.ClientKeyData = string(dataKey)
		}

		s.ClientCertificateDataD = ""
		s.ClientKeyDataD = ""
		s.Token = ""
	case len(s.ClientCertificateData) > 0 && len(s.ClientKeyData) > 0:
		s.ClientCertificateDataD = ""
		s.ClientKeyDataD = ""
		s.Token = ""
	case len(s.Token) > 0:
		s.ClientCertificateDataD = ""
		s.ClientKeyDataD = ""
		s.ClientCertificateData = ""
		s.ClientKeyData = ""
	default:
		return "", errors.New("cannot define valid secret format")
	}

	secretByte, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(secretByte), nil
}

// FormatResponse formats the response for kubectl
func FormatResponse(res *Response) (string, error) {
	err := defaults.Set(res)
	if err != nil {
		return "", err
	}
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(jsonResponse), nil
}
