package service

import "testing"

func TestGenerateGeminiContentSessionHash_StableAcrossTurns(t *testing.T) {
	const prefix = `"tools":[{"functionDeclarations":[{"name":"lookup","parameters":{"type":"object"}}]}],"systemInstruction":{"parts":[{"text":"You are Aria."}]}`
	const first = `{"role":"user","parts":[{"text":"[Story summary] Kim meets Aria."}]},{"role":"model","parts":[{"text":"Noted."}]}`
	turn1 := `{` + prefix + `,"contents":[` + first + `,{"role":"user","parts":[{"text":"guide v1"}]},{"role":"user","parts":[{"text":"msg1"},{"text":"notes v1"}]}]}`
	appended := `{` + prefix + `,"contents":[` + first + `,{"role":"user","parts":[{"text":"guide v1"}]},{"role":"user","parts":[{"text":"msg1"},{"text":"notes v1"}]},{"role":"model","parts":[{"functionCall":{"name":"lookup","args":{}},"thoughtSignature":"c2ln"}]},{"role":"user","parts":[{"functionResponse":{"name":"lookup","response":{}}}]}]}`
	turnContext := `{` + prefix + `,"contents":[` + first + `,{"role":"user","parts":[{"text":"msg1"}]},{"role":"model","parts":[{"text":"reply1"}]},{"role":"user","parts":[{"text":"guide v2"}]},{"role":"user","parts":[{"text":"msg2"},{"text":"notes v2"}]}]}`
	edited := `{` + prefix + `,"contents":[` + first + `,{"role":"user","parts":[{"text":"guide v1"}]},{"role":"user","parts":[{"text":"msg1 edited"}]}],"generationConfig":{"temperature":0.2}}`

	want := GenerateGeminiContentSessionHash(7, "gemini-4-argon", []byte(turn1))
	if want == "" {
		t.Fatal("expected non-empty hash")
	}
	for name, body := range map[string]string{"appended": appended, "turn_context": turnContext, "edited_last": edited} {
		if got := GenerateGeminiContentSessionHash(7, "gemini-4-argon", []byte(body)); got != want {
			t.Errorf("%s: hash %q, want %q", name, got, want)
		}
	}
}

func TestGenerateGeminiContentSessionHash_SeedFormat(t *testing.T) {
	body := `{"tools":[{"googleSearch":{}}],"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`
	want := DeriveSessionHashFromSeed(`gemini_cs_key=3|model=gemini-3.8-flash|tools=[{"googleSearch":{}}]|system={"parts":[{"text":"sys"}]}|first_user=[{"text":"hi"}]`)
	if got := GenerateGeminiContentSessionHash(3, " gemini-3.8-flash ", []byte(body)); got != want {
		t.Fatalf("hash %q, want %q", got, want)
	}
}

func TestGenerateGeminiContentSessionHash_DistinguishesEachSeedField(t *testing.T) {
	const (
		tools    = `"tools":[{"functionDeclarations":[{"name":"a"}]}],`
		system   = `"systemInstruction":{"parts":[{"text":"sys A"}]},`
		first    = `{"role":"user","parts":[{"text":"first A"}]}`
		trailing = `,{"role":"model","parts":[{"text":"reply"}]},{"role":"user","parts":[{"text":"tail"}]}`
	)
	base := `{` + tools + system + `"contents":[` + first + trailing + `]}`
	h := GenerateGeminiContentSessionHash(1, "m1", []byte(base))
	cases := map[string]string{
		"other_key":    GenerateGeminiContentSessionHash(2, "m1", []byte(base)),
		"other_model":  GenerateGeminiContentSessionHash(1, "m2", []byte(base)),
		"no_model":     GenerateGeminiContentSessionHash(1, "", []byte(base)),
		"other_tools":  GenerateGeminiContentSessionHash(1, "m1", []byte(`{"tools":[{"functionDeclarations":[{"name":"b"}]}],`+system+`"contents":[`+first+trailing+`]}`)),
		"no_tools":     GenerateGeminiContentSessionHash(1, "m1", []byte(`{`+system+`"contents":[`+first+trailing+`]}`)),
		"other_system": GenerateGeminiContentSessionHash(1, "m1", []byte(`{`+tools+`"systemInstruction":{"parts":[{"text":"sys B"}]},"contents":[`+first+trailing+`]}`)),
		"no_system":    GenerateGeminiContentSessionHash(1, "m1", []byte(`{`+tools+`"contents":[`+first+trailing+`]}`)),
		"other_first":  GenerateGeminiContentSessionHash(1, "m1", []byte(`{`+tools+system+`"contents":[{"role":"user","parts":[{"text":"first B"}]}`+trailing+`]}`)),
		"inline_first": GenerateGeminiContentSessionHash(1, "m1", []byte(`{`+tools+system+`"contents":[{"role":"user","parts":[{"text":"first A"},{"inlineData":{"mimeType":"image/png","data":"iVBORw0KGgo="}}]}`+trailing+`]}`)),
	}
	for name, got := range cases {
		if got == "" {
			t.Errorf("%s: expected non-empty hash", name)
		}
		if got == h {
			t.Errorf("%s: hash should differ from base", name)
		}
	}
}

func TestGenerateGeminiContentSessionHash_FirstUserSkipsLeadingModelContent(t *testing.T) {
	body := func(firstUser string) []byte {
		return []byte(`{"systemInstruction":{"parts":[{"text":"You are Aria."}]},"contents":[{"role":"model","parts":[{"text":"Hi! I'm Aria."}]},{"role":"user","parts":[{"text":"` + firstUser + `"}]},{"role":"model","parts":[{"text":"reply"}]},{"role":"user","parts":[{"text":"next"}]}]}`)
	}
	kim := GenerateGeminiContentSessionHash(1, "m", body("I'm Kim."))
	lee := GenerateGeminiContentSessionHash(1, "m", body("I'm Lee."))
	if kim == "" || lee == "" {
		t.Fatal("expected non-empty hashes")
	}
	if kim == lee {
		t.Error("conversations sharing only the leading model greeting should not share a hash")
	}

	userFirst := `{"systemInstruction":{"parts":[{"text":"You are Aria."}]},"contents":[{"role":"user","parts":[{"text":"I'm Kim."}]}]}`
	if got := GenerateGeminiContentSessionHash(1, "m", []byte(userFirst)); got != kim {
		t.Errorf("leading model content should be skipped: hash %q, want %q", got, kim)
	}
}

func TestGenerateGeminiContentSessionHash_RoleOmittedCountsAsUser(t *testing.T) {
	withRole := `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`
	withoutRole := `{"contents":[{"parts":[{"text":"hi"}]}]}`
	if GenerateGeminiContentSessionHash(1, "m", []byte(withoutRole)) == "" {
		t.Fatal("expected non-empty hash when role is omitted")
	}
	if GenerateGeminiContentSessionHash(1, "m", []byte(withRole)) != GenerateGeminiContentSessionHash(1, "m", []byte(withoutRole)) {
		t.Error("omitted role and explicit user role should share a hash")
	}
}

func TestGenerateGeminiContentSessionHash_NonTextPartAnchors(t *testing.T) {
	inlineOnly := `{"contents":[{"role":"user","parts":[{"inlineData":{"mimeType":"image/png","data":"iVBORw0KGgo="}}]}]}`
	blankThenInline := `{"contents":[{"role":"user","parts":[{"text":" "},{"fileData":{"fileUri":"gs://b/f.pdf"}}]}]}`
	for name, body := range map[string]string{"inline_only": inlineOnly, "blank_then_file": blankThenInline} {
		if GenerateGeminiContentSessionHash(1, "m", []byte(body)) == "" {
			t.Errorf("%s: expected non-empty hash", name)
		}
	}
}

func TestGenerateGeminiContentSessionHash_NormalizesJSONFormatting(t *testing.T) {
	compact := `{"tools":[{"googleSearch":{}}],"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[{"text":"first"}]}]}`
	reordered := `{ "contents" : [ { "parts" : [ { "text" : "first" } ], "role" : "user" } ], "systemInstruction" : { "parts" : [ { "text" : "sys" } ] }, "tools" : [ { "googleSearch" : { } } ] }`
	emptyToolsBase := `{"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[{"text":"first"}]}]}`
	emptyTools := `{"tools":[],"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[{"text":"first"}]}]}`
	if GenerateGeminiContentSessionHash(1, "m", []byte(emptyTools)) != GenerateGeminiContentSessionHash(1, "m", []byte(emptyToolsBase)) {
		t.Error("empty tools should hash like missing tools")
	}
	snakeCase := `{"tools":[{"googleSearch":{}}],"system_instruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[{"text":"first"}]}]}`
	bothSystems := `{"tools":[{"googleSearch":{}}],"system_instruction":{"parts":[{"text":"other"}]},"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[{"text":"first"}]}]}`

	want := GenerateGeminiContentSessionHash(1, "m", []byte(compact))
	for name, body := range map[string]string{"reordered": reordered, "snake_case_system": snakeCase, "camel_case_system_wins": bothSystems} {
		if got := GenerateGeminiContentSessionHash(1, "m", []byte(body)); got != want {
			t.Errorf("%s: hash %q, want %q", name, got, want)
		}
	}
}

func TestGenerateGeminiContentSessionHash_EmptyWithoutFirstUserContent(t *testing.T) {
	for name, body := range map[string]string{
		"empty_body":       ``,
		"non_object_body":  `[1,2]`,
		"no_contents":      `{"systemInstruction":{"parts":[{"text":"sys"}]}}`,
		"contents_object":  `{"contents":{"role":"user","parts":[{"text":"hi"}]}}`,
		"empty_contents":   `{"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[]}`,
		"only_model":       `{"contents":[{"role":"model","parts":[{"text":"Hi!"}]}]}`,
		"empty_parts":      `{"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user","parts":[]}]}`,
		"missing_parts":    `{"systemInstruction":{"parts":[{"text":"sys"}]},"contents":[{"role":"user"}]}`,
		"blank_text":       `{"contents":[{"role":"user","parts":[{"text":"   "}]}]}`,
		"empty_part":       `{"contents":[{"role":"user","parts":[{}]}]}`,
		"null_part":        `{"contents":[{"role":"user","parts":[null]}]}`,
		"non_object_first": `{"contents":["hello",{"role":"user","parts":[{"text":"hi"}]}]}`,
	} {
		if got := GenerateGeminiContentSessionHash(1, "m", []byte(body)); got != "" {
			t.Errorf("%s: expected empty hash, got %q", name, got)
		}
	}
}
