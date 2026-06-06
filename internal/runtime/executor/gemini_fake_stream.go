package executor

import (
	"context"
	"net/http"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
)

const geminiFakeStreamSuffix = "[假流]"

func geminiFakeStreamEnabled(cfg *config.Config) bool {
	return cfg != nil && cfg.Streaming.GeminiFakeStream
}

func geminiFakeStreamModel(model string) bool {
	return strings.HasSuffix(model, geminiFakeStreamSuffix) || strings.Contains(model, "[假流]")
}

func stripGeminiFakeStreamSuffix(model string) string {
	idx := strings.Index(model, "[假流]")
	if idx >= 0 {
		return model[:idx]
	}
	return strings.TrimSuffix(model, geminiFakeStreamSuffix)
}

func geminiFakeStreamHeaders(headers http.Header) http.Header {
	out := headers.Clone()
	if out == nil {
		out = make(http.Header)
	}
	out.Set("Content-Type", "text/event-stream")
	out.Set("Cache-Control", "no-cache")
	return out
}

func geminiFakeStreamResult(ctx context.Context, headers http.Header, from, to sdktranslator.Format, model string, originalRequestRawJSON, requestRawJSON, rawJSON []byte) *cliproxyexecutor.StreamResult {
	out := make(chan cliproxyexecutor.StreamChunk)
	go func() {
		defer close(out)
		var param any
		emit := func(payload []byte) bool {
			if len(payload) == 0 {
				return true
			}
			select {
			case out <- cliproxyexecutor.StreamChunk{Payload: payload}:
				return true
			case <-ctx.Done():
				return false
			}
		}

		for _, chunk := range sdktranslator.TranslateStream(ctx, to, from, model, originalRequestRawJSON, requestRawJSON, rawJSON, &param) {
			if !emit(chunk) {
				return
			}
		}
		for _, chunk := range sdktranslator.TranslateStream(ctx, to, from, model, originalRequestRawJSON, requestRawJSON, []byte("[DONE]"), &param) {
			if !emit(chunk) {
				return
			}
		}
	}()
	return &cliproxyexecutor.StreamResult{Headers: geminiFakeStreamHeaders(headers), Chunks: out}
}
