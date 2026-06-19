package executor

import (
	"bytes"
	"context"
	"testing"
	"time"

	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

func TestGeminiFakeStreamResultWithHeartbeatEmitsBeforeCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	unblock := make(chan struct{})
	result := geminiFakeStreamResultWithHeartbeat(ctx, func() (*cliproxyexecutor.StreamResult, error) {
		<-unblock
		chunks := make(chan cliproxyexecutor.StreamChunk, 1)
		chunks <- cliproxyexecutor.StreamChunk{Payload: []byte(`{"done":true}`)}
		close(chunks)
		return &cliproxyexecutor.StreamResult{Chunks: chunks}, nil
	})

	select {
	case chunk := <-result.Chunks:
		if !bytes.Equal(chunk.Payload, geminiFakeStreamHeartbeatChunk()) {
			t.Fatalf("first chunk = %q, want heartbeat", chunk.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for heartbeat before fake stream completion")
	}

	close(unblock)

	select {
	case chunk := <-result.Chunks:
		if string(chunk.Payload) != `{"done":true}` {
			t.Fatalf("payload chunk = %q, want final payload", chunk.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for final payload")
	}
}

func TestNormalizeGeminiFakeStreamModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		model      string
		want       string
		fakeStream bool
	}{
		{name: "fake stream variant", model: "gemini-3.1-pro-preview[假流]", want: "gemini-3.1-pro-preview", fakeStream: true},
		{name: "fake stream variant with whitespace", model: " gemini-2.5-pro[假流] ", want: "gemini-2.5-pro", fakeStream: true},
		{name: "normal model", model: "gemini-3.1-pro-preview", want: "gemini-3.1-pro-preview", fakeStream: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, fakeStream := normalizeGeminiFakeStreamModel(tc.model)
			if got != tc.want || fakeStream != tc.fakeStream {
				t.Fatalf("normalizeGeminiFakeStreamModel(%q) = (%q, %t), want (%q, %t)", tc.model, got, fakeStream, tc.want, tc.fakeStream)
			}
		})
	}
}
