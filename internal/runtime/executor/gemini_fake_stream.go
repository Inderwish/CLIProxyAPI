package executor

import (
	"context"
	"net/http"
	"strings"
	"time"

	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
)

const geminiFakeStreamSuffix = "[假流]"

const geminiFakeStreamHeartbeatInterval = 5 * time.Second

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

func normalizeGeminiFakeStreamModel(model string) (string, bool) {
	if !geminiFakeStreamModel(model) {
		return model, false
	}
	return strings.TrimSpace(stripGeminiFakeStreamSuffix(model)), true
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

func geminiFakeStreamHeartbeatChunk() []byte {
	return []byte(": keep-alive\n\n")
}

func geminiFakeStreamResultWithHeartbeat(ctx context.Context, execute func() (*cliproxyexecutor.StreamResult, error)) *cliproxyexecutor.StreamResult {
	out := make(chan cliproxyexecutor.StreamChunk)
	go func() {
		defer close(out)

		emit := func(chunk cliproxyexecutor.StreamChunk) bool {
			if chunk.Err == nil && len(chunk.Payload) == 0 {
				return true
			}
			select {
			case out <- chunk:
				return true
			case <-ctx.Done():
				return false
			}
		}

		if !emit(cliproxyexecutor.StreamChunk{Payload: geminiFakeStreamHeartbeatChunk()}) {
			return
		}

		resultDone := make(chan struct{})
		var result *cliproxyexecutor.StreamResult
		var resultErr error
		go func() {
			defer close(resultDone)
			result, resultErr = execute()
		}()

		ticker := time.NewTicker(geminiFakeStreamHeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !emit(cliproxyexecutor.StreamChunk{Payload: geminiFakeStreamHeartbeatChunk()}) {
					return
				}
			case <-resultDone:
				if resultErr != nil {
					emit(cliproxyexecutor.StreamChunk{Err: resultErr})
					return
				}
				if result == nil {
					return
				}
				if result.Chunks == nil {
					return
				}
				for chunk := range result.Chunks {
					if !emit(chunk) {
						return
					}
				}
				return
			}
		}
	}()
	return &cliproxyexecutor.StreamResult{Headers: geminiFakeStreamHeaders(nil), Chunks: out}
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
