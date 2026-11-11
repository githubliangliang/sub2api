# Measured prerequisites

- Gemini isGeminiThoughtPart: 4 lines (bool type assertion); complete upstream gemini_compat_thinking.go: 320 lines. Only the helper is needed for F08.
- OAuth WS search activation: new source helper 60 lines; local AllItems, openAIToolsContainWebSearch and gjsonToolsContainWebSearch already exist.
- Codex page context: source helper 123 lines; existing mayContainCodexCallOutputCandidate and hasUniqueJSONMembers are reusable.
- OpenAI cooldown: local newOpenAIStreamFailoverError has no canonical-model argument or terminal model side effects. Source WithModel-only tests do not compile. Actual local HTTP/four reader probes compile and fail desired behavior. S04 must include the consumer chain.
- Native WS ingress: local public function returns anonymous error; source defer uses named returnErr. S01 must adapt return flow, not paste defer verbatim.
- SQLite: source index uses CONCURRENTLY/INCLUDE; temporary real SQLite proof uses third actual_cost column with same repository results. No migration has been added to product.
- Fast free-input fix in PR8606 is inside newly introduced CacheCreation12hPricePerToken handling. Existing local no-TTL Fast negative control passes.
