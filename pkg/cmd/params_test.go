package cmd

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/hpinc/tcli/pkg/parser"
)

func TestGetGlobalTraceFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	r := &parser.Root{}

	g := getGlobal(fs, r)

	if g.Trace != false {
		t.Errorf("Expected trace to default to false")
	}

	// Test setting trace to true via flag
	err := fs.Parse([]string{"-trace"})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if !g.Trace {
		t.Errorf("Expected Trace to be true after parsing -trace flag")
	}
}

func TestSecretStringValueRedactsDefaultAndRetainsValue(t *testing.T) {
	token := "cached-sensitive-token"
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var output bytes.Buffer
	fs.SetOutput(&output)
	fs.Var(secretStringValue{value: &token}, JwtParam, "bearer token")

	if got := fs.Lookup(JwtParam).DefValue; got != "******" {
		t.Fatalf("JWT default = %q, want redacted", got)
	}
	fs.PrintDefaults()
	if strings.Contains(output.String(), token) {
		t.Fatalf("help output exposed JWT: %s", output.String())
	}
	if !strings.Contains(output.String(), "******") {
		t.Fatalf("help output omitted redacted JWT value: %s", output.String())
	}
	if token != "cached-sensitive-token" {
		t.Fatalf("registering the flag changed the actual token to %q", token)
	}

	if err := fs.Parse([]string{"-jwt", "override-token"}); err != nil {
		t.Fatalf("parse JWT override: %v", err)
	}
	if token != "override-token" {
		t.Fatalf("parsed JWT = %q, want override-token", token)
	}
}
