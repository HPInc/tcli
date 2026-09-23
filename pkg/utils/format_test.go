package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hpinc/tcli/pkg/common"
)

func TestDoFormat_Trace(t *testing.T) {
	jsonData := []byte(`{"id":1, "name":"test"}`)
	values := make(common.Values)

	t.Run("WithoutTrace", func(t *testing.T) {
		oldStdout := os.Stdout
		oldStderr := os.Stderr

		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()

		os.Stdout = wOut
		os.Stderr = wErr

		// Execute without trace
		DoFormat(jsonData, values, ".", false)

		wOut.Close()
		wErr.Close()

		os.Stdout = oldStdout
		os.Stderr = oldStderr

		var bufOut bytes.Buffer
		io.Copy(&bufOut, rOut)
		outStr := bufOut.String()

		var bufErr bytes.Buffer
		io.Copy(&bufErr, rErr)
		errStr := bufErr.String()

		if !strings.Contains(outStr, `"id":1`) {
			t.Errorf("Expected stdout to contain JSON, got: %s", outStr)
		}
		if len(errStr) > 0 {
			t.Errorf("Expected empty stderr without trace, got: %s", errStr)
		}
	})

	t.Run("WithTrace", func(t *testing.T) {
		oldStdout := os.Stdout
		oldStderr := os.Stderr

		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()

		os.Stdout = wOut
		os.Stderr = wErr

		// Execute with trace
		DoFormat(jsonData, values, ".", true)

		wOut.Close()
		wErr.Close()

		os.Stdout = oldStdout
		os.Stderr = oldStderr

		var bufOut bytes.Buffer
		io.Copy(&bufOut, rOut)
		outStr := bufOut.String()

		var bufErr bytes.Buffer
		io.Copy(&bufErr, rErr)
		errStr := bufErr.String()

		if !strings.Contains(outStr, `"id":1`) {
			t.Errorf("Expected stdout to contain JSON, got: %s", outStr)
		}
		if !strings.Contains(errStr, `"id":1`) {
			t.Errorf("Expected stderr to contain JSON with trace, got: %s", errStr)
		}
	})
}
