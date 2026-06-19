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
