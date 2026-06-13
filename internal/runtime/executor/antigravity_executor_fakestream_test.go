package executor

import (
	"bytes"
	"testing"

	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestAntigravityOpenAIChatFakeStreamChunksAreRawJSON(t *testing.T) {
	payload := []byte(`{"id":"chatcmpl-test","created":123,"model":"gemini-response","choices":[{"message":{"content":"hello","reasoning_content":"why"},"finish_reason":"stop"}]}`)

	chunks := antigravityFinalFakeStreamChunks(sdktranslator.FormatOpenAI, "gemini-request", payload)
	if len(chunks) != 2 {
		t.Fatalf("chunk count = %d, want 2", len(chunks))
	}

	for i, chunk := range chunks {
		if bytes.HasPrefix(chunk, []byte("data:")) {
			t.Fatalf("chunk %d has SSE data prefix: %q", i, chunk)
		}
		if bytes.Contains(chunk, []byte("[DONE]")) {
			t.Fatalf("chunk %d contains DONE marker: %q", i, chunk)
		}
		if !gjson.ValidBytes(chunk) {
			t.Fatalf("chunk %d is not valid JSON: %q", i, chunk)
		}
	}

	if got := gjson.GetBytes(chunks[0], "choices.0.delta.content").String(); got != "hello" {
		t.Fatalf("content delta = %q, want %q", got, "hello")
	}
	if got := gjson.GetBytes(chunks[0], "choices.0.delta.reasoning_content").String(); got != "why" {
		t.Fatalf("reasoning delta = %q, want %q", got, "why")
	}
	if got := gjson.GetBytes(chunks[1], "choices.0.finish_reason").String(); got != "stop" {
		t.Fatalf("finish_reason = %q, want %q", got, "stop")
	}
}

func TestAntigravityOpenAIFakeStreamHeartbeatChunkIsRawJSON(t *testing.T) {
	chunk := antigravityFakeStreamHeartbeatChunk(sdktranslator.FormatOpenAI, "gemini-3.5-flash-low")

	if len(chunk) == 0 {
		t.Fatal("heartbeat chunk is empty")
	}
	if bytes.HasPrefix(chunk, []byte("data:")) {
		t.Fatalf("heartbeat has SSE data prefix: %q", chunk)
	}
	if !gjson.ValidBytes(chunk) {
		t.Fatalf("heartbeat is not valid JSON: %q", chunk)
	}
	if got := gjson.GetBytes(chunk, "choices.0.delta").Raw; got != "{}" {
		t.Fatalf("heartbeat delta = %q, want {}", got)
	}
	if got := gjson.GetBytes(chunk, "choices.0.finish_reason").Raw; got != "null" {
		t.Fatalf("heartbeat finish_reason = %q, want null", got)
	}
}
