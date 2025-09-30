package provider

import (
	"os"
)

const defaultEndpoint = "https://orchestrator.fabric-testbed.net"

func getenvOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
