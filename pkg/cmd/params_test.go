package cmd

import (
	"flag"
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
