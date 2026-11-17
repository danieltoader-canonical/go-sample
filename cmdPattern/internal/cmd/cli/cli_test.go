package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSetup(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{name: "help", args: []string{"--help"}, want: "consumer"},
		{name: "version", args: []string{"--version"}, want: "cli version"},
		{name: "short version", args: []string{"command", "get-version", "-s"}, want: "1.0.0+h.-t.0"},
		{name: "consumer help", args: []string{"consumer", "--help"}, want: "sample"},
		{name: "sample consumer", args: []string{"consumer", "sample"}, want: "Sample consumer"},
		{name: "sample arguments", args: []string{"consumer", "sample", "unexpected"}, wantErr: "unknown command"},
		{name: "unknown command", args: []string{"unknown"}, wantErr: "unknown command"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Setup(context.Background())
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got %v, want %q", err, tt.wantErr)
				}
			} else if err != nil || !strings.Contains(out.String(), tt.want) {
				t.Fatalf("err=%v output=%q, want %q", err, out.String(), tt.want)
			}
		})
	}
}
