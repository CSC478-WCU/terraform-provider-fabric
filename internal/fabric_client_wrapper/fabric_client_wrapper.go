package fabric_client_wrapper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	openapi "github.com/csc478-wcu/fabric-orchestrator-go-client"
)

// FabricClientWrapper wraps the generated client to handle content-type issues
type FabricClientWrapper struct {
	*openapi.APIClient
}

// NewFabricClientWrapper creates a new wrapper around the fabric client
func NewFabricClientWrapper(cfg *openapi.Configuration) *FabricClientWrapper {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}
	originalTransport := cfg.HTTPClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	cfg.HTTPClient.Transport = &contentTypeFixTransport{
		wrapped: originalTransport,
	}

	return &FabricClientWrapper{
		APIClient: openapi.NewAPIClient(cfg),
	}
}

// contentTypeFixTransport wraps an http.RoundTripper to fix content-type issues
type contentTypeFixTransport struct {
	wrapped http.RoundTripper
}

func (t *contentTypeFixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.wrapped.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode == 200 &&
		strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		resp.Body.Close()

		if isJSON(bodyBytes) {
			resp.Header.Set("Content-Type", "application/json")
		}
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	return resp, nil
}

func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
