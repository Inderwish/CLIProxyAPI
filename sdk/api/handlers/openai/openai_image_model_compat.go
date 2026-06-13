package openai

import (
	"strings"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func compatGPTImage2RequestForChat(cfg *internalconfig.SDKConfig, rawJSON []byte) ([]byte, bool) {
	return compatGPTImage2RequestForNonImageEndpoint(cfg, rawJSON)
}

func compatGPTImage2RequestForResponses(cfg *internalconfig.SDKConfig, rawJSON []byte) ([]byte, bool) {
	return compatGPTImage2RequestForNonImageEndpoint(cfg, rawJSON)
}

func compatGPTImage2RequestForNonImageEndpoint(cfg *internalconfig.SDKConfig, rawJSON []byte) ([]byte, bool) {
	if imageGenerationDisabledForNonImageEndpoint(cfg) {
		return rawJSON, false
	}
	requestedModel := strings.TrimSpace(gjson.GetBytes(rawJSON, "model").String())
	if !isDefaultImagesToolModel(requestedModel) {
		return rawJSON, false
	}

	tool := gptImage2CompatTool(rawJSON, requestedModel)
	out := rawJSON
	out, _ = sjson.SetBytes(out, "model", gptImage2CompatMainModel(cfg, requestedModel))
	out, _ = sjson.SetRawBytes(out, "tools", []byte(`[]`))
	out, _ = sjson.SetRawBytes(out, "tools.-1", tool)
	out, _ = sjson.SetRawBytes(out, "tool_choice", []byte(`{"type":"image_generation"}`))
	return out, true
}

func imageGenerationDisabledForNonImageEndpoint(cfg *internalconfig.SDKConfig) bool {
	return cfg != nil &&
		(cfg.DisableImageGeneration == internalconfig.DisableImageGenerationAll ||
			cfg.DisableImageGeneration == internalconfig.DisableImageGenerationChat)
}

func gptImage2CompatMainModel(cfg *internalconfig.SDKConfig, requestedModel string) string {
	mainModel := defaultImagesMainModel
	if cfg != nil {
		if configured := strings.TrimSpace(cfg.GPTImage2BaseModel); strings.HasPrefix(strings.ToLower(configured), "gpt-") {
			mainModel = configured
		}
	}

	prefix, _ := imagesModelParts(requestedModel)
	if prefix == "" {
		return mainModel
	}
	return prefix + "/" + mainModel
}

func gptImage2CompatTool(rawJSON []byte, requestedModel string) []byte {
	tool := []byte(`{"type":"image_generation"}`)
	tool, _ = sjson.SetBytes(tool, "model", requestedModel)

	for _, field := range []string{"size", "quality", "background", "output_format", "input_fidelity", "moderation"} {
		if value := strings.TrimSpace(gjson.GetBytes(rawJSON, field).String()); value != "" {
			tool, _ = sjson.SetBytes(tool, field, value)
		}
	}
	for _, field := range []string{"output_compression", "partial_images"} {
		value := gjson.GetBytes(rawJSON, field)
		if value.Exists() && value.Type == gjson.Number {
			tool, _ = sjson.SetBytes(tool, field, value.Int())
		}
	}

	return tool
}
