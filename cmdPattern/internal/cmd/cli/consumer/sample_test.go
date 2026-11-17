package consumer

import (
	"bytes"
	"errors"
	"testing"
)

func TestSampleConsumer(t *testing.T) {
	cmd := sampleConsumer()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Sample consumer\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestSampleConsumerWriteError(t *testing.T) {
	failure := errors.New("write failed")
	cmd := sampleConsumer()
	cmd.SetOut(sampleFailWriter{failure})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); !errors.Is(err, failure) {
		t.Fatalf("got %v, want %v", err, failure)
	}
}

type sampleFailWriter struct{ err error }

func (w sampleFailWriter) Write([]byte) (int, error) { return 0, w.err }
