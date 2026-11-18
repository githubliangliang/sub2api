package service

import "github.com/tidwall/gjson"

// Match the selected provider compatibility rules without using a full regexp
// compiler: preserve escaped literals, character classes and Python named groups.
func hasUnsupportedOpenAIRegexPattern(pattern string) bool {
	characterClass := false
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			if i+1 >= len(pattern) {
				return true
			}
			escaped := pattern[i+1]
			if (escaped >= 'a' && escaped <= 'z' || escaped >= 'A' && escaped <= 'Z') && !pythonRegexEscapeAllowed(escaped) {
				return true
			}
			i++
		case '[':
			characterClass = true
		case ']':
			characterClass = false
		case '(':
			if characterClass || i+1 >= len(pattern) || pattern[i+1] != '?' {
				continue
			}
			if i+2 >= len(pattern) {
				return true
			}
			switch pattern[i+2] {
			case '=', '!', '<', '\'':
				return true
			}
		}
	}
	return false
}

func pythonRegexEscapeAllowed(ch byte) bool {
	switch ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', 'd', 'D', 's', 'S', 'w', 'W', 'A', 'Z', 'B', 'N', 'u', 'U', 'x':
		return true
	}
	return false
}

type openAISchemaMember struct{ key, value gjson.Result }

func openAISchemaMembers(object gjson.Result) []openAISchemaMember {
	var members []openAISchemaMember
	object.ForEach(func(key, value gjson.Result) bool {
		members = append(members, openAISchemaMember{key, value})
		return true
	})
	return members
}

// Delete consecutive runs together so adjacent removals cannot consume the
// same comma. Offsets always refer to the original body, including nested edits.
func collectOpenAISchemaMemberDeletions(body []byte, members []openAISchemaMember, deleted []bool, hits *[]openAIResponsesToolSchemaNullType) {
	for i := 0; i < len(members); i++ {
		if !deleted[i] {
			continue
		}
		last := i
		for last+1 < len(members) && deleted[last+1] {
			last++
		}
		start, end := members[i].key.Index, members[last].value.Index+len(members[last].value.Raw)
		if start <= 0 || end > len(body) {
			i = last
			continue
		}
		if last+1 < len(members) {
			for end < len(body) && (body[end] == ' ' || body[end] == '\t' || body[end] == '\r' || body[end] == '\n') {
				end++
			}
			if end < len(body) && body[end] == ',' {
				end++
			}
		} else if i > 0 {
			previous := start - 1
			for previous >= 0 && (body[previous] == ' ' || body[previous] == '\t' || body[previous] == '\r' || body[previous] == '\n') {
				previous--
			}
			if previous >= 0 && body[previous] == ',' {
				start = previous
			}
		}
		*hits = append(*hits, openAIResponsesToolSchemaNullType{offset: start, length: end - start})
		i = last
	}
}

// Walk only schema-valued keywords. default, enum, const and examples contain
// instance data and must keep fields and strings named pattern unchanged.
func collectOpenAISchemaRegexEdits(body []byte, root gjson.Result, hits *[]openAIResponsesToolSchemaNullType) {
	pending := []gjson.Result{root}
	for len(pending) > 0 {
		schema := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if schema.IsArray() {
			pending = append(pending, schema.Array()...)
			continue
		}
		if !schema.IsObject() {
			continue
		}
		members := openAISchemaMembers(schema)
		deleted := make([]bool, len(members))
		lostPatternProperty := false
		for i, member := range members {
			switch member.key.String() {
			case "required":
				deleted[i] = member.value.Raw == "null"
			case "pattern":
				deleted[i] = member.value.Type == gjson.String && hasUnsupportedOpenAIRegexPattern(member.value.String())
			case "patternProperties":
				if !member.value.IsObject() {
					continue
				}
				patterns := openAISchemaMembers(member.value)
				removed := make([]bool, len(patterns))
				for n, pattern := range patterns {
					removed[n] = hasUnsupportedOpenAIRegexPattern(pattern.key.String())
					if removed[n] {
						lostPatternProperty = true
					} else {
						pending = append(pending, pattern.value)
					}
				}
				collectOpenAISchemaMemberDeletions(body, patterns, removed, hits)
			}
		}
		collectOpenAISchemaMemberDeletions(body, members, deleted, hits)
		for i, member := range members {
			if deleted[i] {
				continue
			}
			if member.key.String() == "additionalProperties" && lostPatternProperty {
				if member.value.Raw != "true" {
					*hits = append(*hits, openAIResponsesToolSchemaNullType{offset: member.value.Index, length: len(member.value.Raw), replacement: "true"})
				}
				continue // Its nested constraints are superseded by the enclosing edit.
			}
			switch member.key.String() {
			case "properties", "$defs", "definitions", "dependentSchemas", "dependencies":
				member.value.ForEach(func(_, child gjson.Result) bool { pending = append(pending, child); return true })
			case "items", "prefixItems", "allOf", "anyOf", "oneOf", "additionalProperties", "additionalItems", "contains", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems", "contentSchema":
				pending = append(pending, member.value)
			}
		}
	}
}
