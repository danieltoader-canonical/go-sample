package command

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "full", args: []string{}, want: fmt.Sprintf("Application Name: default \nVersion: 1.0.0\nHash: \nBuild Time: \nGo version: %s\n", runtime.Version())},
		{name: "short", args: []string{"-s"}, want: "1.0.0+h.-t.0\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := getVersionCommand()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Fatalf("got %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestVersionCommandWriteError(t *testing.T) {
	failure := errors.New("write failed")
	cmd := getVersionCommand()
	cmd.SetOut(failWriter{failure})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); !errors.Is(err, failure) {
		t.Fatalf("got %v, want %v", err, failure)
	}
}

type failWriter struct{ err error }

func (w failWriter) Write([]byte) (int, error) { return 0, w.err }
