package openai

import (
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
)

func TestCompatGPTImage2RequestForChatForcesImageGenerationTool(t *testing.T) {
	raw := []byte(`{"model":"gpt-image-2","messages":[{"role":"user","content":"draw a fox"}],"stream":false,"size":"1024x1024","quality":"high","partial_images":2}`)

	out, ok := compatGPTImage2RequestForChat(nil, raw)
	if !ok {
		t.Fatal("compatGPTImage2RequestForChat() did not convert gpt-image-2 request")
	}

	if got := gjson.GetBytes(out, "model").String(); got != defaultImagesMainModel {
		t.Fatalf("model = %q, want %q", got, defaultImagesMainModel)
	}
	if got := gjson.GetBytes(out, "tools.0.type").String(); got != "image_generation" {
		t.Fatalf("tools.0.type = %q, want image_generation", got)
	}
	if got := gjson.GetBytes(out, "tools.0.model").String(); got != "gpt-image-2" {
		t.Fatalf("tools.0.model = %q, want gpt-image-2", got)
	}
	if got := gjson.GetBytes(out, "tools.0.size").String(); got != "1024x1024" {
		t.Fatalf("tools.0.size = %q, want 1024x1024", got)
	}
	if got := gjson.GetBytes(out, "tools.0.partial_images").Int(); got != 2 {
		t.Fatalf("tools.0.partial_images = %d, want 2", got)
	}
	if got := gjson.GetBytes(out, "tool_choice.type").String(); got != "image_generation" {
		t.Fatalf("tool_choice.type = %q, want image_generation", got)
	}
}

func TestCompatGPTImage2RequestForResponsesPreservesProviderPrefix(t *testing.T) {
	cfg := &internalconfig.SDKConfig{GPTImage2BaseModel: "gpt-5.4"}
	raw := []byte(`{"model":"codex/gpt-image-2","input":"draw a castle"}`)

	out, ok := compatGPTImage2RequestForResponses(cfg, raw)
	if !ok {
		t.Fatal("compatGPTImage2RequestForResponses() did not convert prefixed gpt-image-2 request")
	}

	if got := gjson.GetBytes(out, "model").String(); got != "codex/gpt-5.4" {
		t.Fatalf("model = %q, want codex/gpt-5.4", got)
	}
	if got := gjson.GetBytes(out, "tools.0.model").String(); got != "codex/gpt-image-2" {
		t.Fatalf("tools.0.model = %q, want codex/gpt-image-2", got)
	}
}

func TestCompatGPTImage2RequestForChatDisabledWhenImageGenerationDisabledForChat(t *testing.T) {
	cfg := &internalconfig.SDKConfig{DisableImageGeneration: internalconfig.DisableImageGenerationChat}
	raw := []byte(`{"model":"gpt-image-2","messages":[{"role":"user","content":"draw"}]}`)

	out, ok := compatGPTImage2RequestForChat(cfg, raw)
	if ok {
		t.Fatal("compatGPTImage2RequestForChat() converted request despite chat image generation being disabled")
	}
	if string(out) != string(raw) {
		t.Fatalf("request mutated despite disabled image generation: %s", string(out))
	}
}
