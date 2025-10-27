package internal

import (
	"os"
	"testing"
)

func TestGetNodeName(t *testing.T) {
	// Table-driven tests
	tests := []struct {
		name          string
		envValue      string
		setEnv        bool
		expectFromEnv bool
	}{
		{name: "env var set", envValue: "node-from-env", setEnv: true, expectFromEnv: true},
		{name: "env var empty -> fallback to hostname", envValue: "", setEnv: false, expectFromEnv: false},
	}

	original := os.Getenv("NODE_NAME")
	defer func() {
		// Restore original env
		_ = os.Setenv("NODE_NAME", original)
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				_ = os.Setenv("NODE_NAME", tt.envValue)
			} else {
				_ = os.Unsetenv("NODE_NAME")
			}

			got, err := GetNodeName()
			if err != nil {
				t.Fatalf("getNodeName returned error: %v", err)
			}

			if tt.expectFromEnv {
				if got != tt.envValue {
					t.Fatalf("expected name from env %q, got %q", tt.envValue, got)
				}
			} else {
				host, _ := os.Hostname()
				if got != host {
					t.Fatalf("expected fallback to os.Hostname %q, got %q", host, got)
				}
			}
		})
	}
}

func TestPingHost_Errors(t *testing.T) {
	// Only test error cases to avoid network and permission dependencies.
	tests := []struct {
		name        string
		destination string
	}{
		{name: "empty destination", destination: ""},
		{name: "invalid hostname", destination: "invalid_host_$$$"},
		{name: "unresolvable TLD", destination: "nonexistent.invalidtld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := PingHost(tt.destination); err == nil {
				t.Fatalf("expected error for destination %q, got nil", tt.destination)
			}
		})
	}
}

func TestGetNodeList_OutOfCluster(t *testing.T) {
	// We expect an error when running outside a Kubernetes cluster.
	nodes, err := GetNodeList()
	if err == nil {
		t.Fatalf("expected error when not running in cluster, got nodes: %v", nodes)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected empty node list on error, got: %v", nodes)
	}
}
