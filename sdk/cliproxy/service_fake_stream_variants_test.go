package cliproxy

import "testing"

func TestApplyFakeStreamModelVariants(t *testing.T) {
	t.Parallel()

	models := []*ModelInfo{
		{
			ID:          "gemini-3.1-pro-preview",
			Object:      "model",
			Type:        "gemini",
			DisplayName: "Gemini 3.1 Pro Preview",
			Name:        "models/gemini-3.1-pro-preview",
			Description: "Gemini 3.1 Pro Preview",
		},
		{
			ID:          "gemini-3.1-flash-lite-preview[假流]",
			Object:      "model",
			Type:        "gemini",
			DisplayName: "Gemini 3.1 Flash Lite Preview [假流]",
			Name:        "models/gemini-3.1-flash-lite-preview[假流]",
		},
	}

	got := applyFakeStreamModelVariants(models)
	ids := make(map[string]*ModelInfo, len(got))
	for _, model := range got {
		if model != nil {
			ids[model.ID] = model
		}
	}

	for _, want := range []string{
		"gemini-3.1-pro-preview",
		"gemini-3.1-pro-preview[假流]",
		"gemini-3.1-flash-lite-preview[假流]",
	} {
		if ids[want] == nil {
			t.Fatalf("missing model variant %q in %#v", want, ids)
		}
	}
	if len(got) != 3 {
		t.Fatalf("model count = %d, want 3; ids=%#v", len(got), ids)
	}
	if gotName := ids["gemini-3.1-pro-preview[假流]"].Name; gotName != "models/gemini-3.1-pro-preview[假流]" {
		t.Fatalf("fake stream name = %q, want models/gemini-3.1-pro-preview[假流]", gotName)
	}
	if gotDisplay := ids["gemini-3.1-pro-preview[假流]"].DisplayName; gotDisplay != "Gemini 3.1 Pro Preview [假流]" {
		t.Fatalf("fake stream display name = %q, want Gemini 3.1 Pro Preview [假流]", gotDisplay)
	}
}

func TestFakeStreamVariantsEnabledForProvider(t *testing.T) {
	t.Parallel()

	if !fakeStreamVariantsEnabledForProvider("gemini-cli") {
		t.Fatalf("gemini-cli should expose fake stream variants")
	}
	if fakeStreamVariantsEnabledForProvider("antigravity") {
		t.Fatalf("antigravity declares fake stream models separately")
	}
	if fakeStreamVariantsEnabledForProvider("gemini") {
		t.Fatalf("gemini should not expose generated fake stream variants")
	}
}
