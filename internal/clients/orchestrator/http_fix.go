package orchestrator

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ctFixTransport struct{ rt http.RoundTripper }

func (t ctFixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if resp != nil && resp.StatusCode == 200 && resp.Header.Get("Content-Type") == "text/html" {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		_ = resp.Body.Close()
		if json.Valid(b) {
			resp.Header.Set("Content-Type", "application/json")
		}
		resp.Body = io.NopCloser(bytes.NewReader(b))
	}
	return resp, nil
}

func withContentTypeFix(c *http.Client) *http.Client {
	if c == nil {
		c = &http.Client{}
	}
	rt := c.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	c.Transport = ctFixTransport{rt: rt}
	return c
}
