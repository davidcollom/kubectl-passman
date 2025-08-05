// Package passman provides utilities for handling Kubernetes ExecCredential responses,
// including validation, normalisation, and formatting of secrets for use with kubectl.
//
// This package defines structures for representing ExecCredential responses and their status,
// and provides functions to validate and process secrets in various formats (token-based or
// certificate-based, including base64-encoded certificates). It also includes error handling
// for invalid secret formats and utilities to marshal responses into JSON for kubectl consumption.
package passman

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/creasty/defaults"
)

// ResponseStatus represents the status field of an ExecCredential response.
type ResponseStatus struct {
	Token                  string `json:"token,omitempty"`
	ClientCertificateData  string `json:"clientCertificateData,omitempty"`
	ClientCertificateDataD string `json:"client-certificate-data,omitempty"` //nolint:tagliatelle
	ClientKeyData          string `json:"clientKeyData,omitempty"`
	ClientKeyDataD         string `json:"client-key-data,omitempty"` //nolint:tagliatelle
}

// Response represents a Kubernetes ExecCredential response.
type Response struct {
	APIVersion string         `default:"client.authentication.k8s.io/v1beta1" json:"apiVersion"`
	Kind       string         `default:"ExecCredential"                       json:"kind"`
	Status     ResponseStatus `                                               json:"status"` //nolint:tagalign
}

// Static Errors...
var (
	ErrInvalidSecretFormat = errors.New("cannot define valid secret format")
)

// FormatValidator validates and normalises the secret format.
func FormatValidator(secret string) (string, error) {
	respStatus := &ResponseStatus{}
	data := []byte(secret)

	err := json.Unmarshal(data, respStatus)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	err = processSecret(respStatus)
	if err != nil {
		return "", fmt.Errorf("failed to process secret: %w", err)
	}

	secretByte, err := json.Marshal(respStatus)
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret: %w", err)
	}

	return string(secretByte), nil
}

// processSecret handles the different secret types and normalises them.
func processSecret(status *ResponseStatus) error {
	switch {
	case status.ClientCertificateDataD != "" && status.ClientKeyDataD != "":
		return processEncodedCertificates(status)
	case status.ClientCertificateData != "" && status.ClientKeyData != "":
		clearCertificateFields(status)
	case status.Token != "":
		clearTokenFields(status)
	default:
		return ErrInvalidSecretFormat
	}

	return nil
}

// processEncodedCertificates decodes base64 encoded certificates.
func processEncodedCertificates(status *ResponseStatus) error {
	dataCrt, errCrt := base64.StdEncoding.DecodeString(status.ClientCertificateDataD)
	if errCrt != nil {
		return fmt.Errorf("failed to decode client certificate data: %w", errCrt)
	}

	dataKey, errKey := base64.StdEncoding.DecodeString(status.ClientKeyDataD)
	if errKey != nil {
		return fmt.Errorf("failed to decode client key data: %w", errKey)
	}

	status.ClientCertificateData = string(dataCrt)
	status.ClientKeyData = string(dataKey)
	clearCertificateFields(status)

	return nil
}

// clearCertificateFields clears certificate-related fields when using certificates.
func clearCertificateFields(status *ResponseStatus) {
	status.ClientCertificateDataD = ""
	status.ClientKeyDataD = ""
	status.Token = ""
}

// clearTokenFields clears token-related fields when using token auth.
func clearTokenFields(status *ResponseStatus) {
	status.ClientCertificateDataD = ""
	status.ClientKeyDataD = ""
	status.ClientCertificateData = ""
	status.ClientKeyData = ""
}

// FormatResponse formats the response for kubectl.
func FormatResponse(res *Response) (string, error) {
	err := defaults.Set(res)
	if err != nil {
		return "", fmt.Errorf("failed to set defaults: %w", err)
	}

	jsonResponse, err := json.Marshal(res)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonResponse), nil
}
