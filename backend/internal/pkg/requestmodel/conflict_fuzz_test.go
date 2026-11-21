package requestmodel

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

func FuzzConflictingFieldKeyEquivalence(f *testing.F) {
	for _, keys := range [][2]string{{"model", "MODEL"}, {"service_tier", "ſervice_tier"}, {"top_k", "top_K"}, {"id", "ID"}, {"\xff", "�"}, {"", ""}, {"a\\b", "a\\b"}} {
		f.Add(keys[0], keys[1])
	}
	f.Fuzz(func(t *testing.T, first, second string) {
		if len(first)+len(second) > 8192 {
			t.Skip()
		}
		a, _ := json.Marshal(first)
		b, _ := json.Marshal(second)
		var da, db string
		_ = json.Unmarshal(a, &da)
		_ = json.Unmarshal(b, &db)
		body := []byte("{" + string(a) + ":1," + string(b) + ":2}")
		_, conflict := ConflictingField("application/json", body)
		if conflict != strings.EqualFold(da, db) {
			t.Fatalf("folded key comparison differs for %q %q", da, db)
		}
		freeform := []byte(`{"arguments":` + string(body) + `}`)
		_, conflict = ConflictingField("application/json", freeform)
		if conflict != (da == db) {
			t.Fatalf("freeform key comparison differs for %q %q", da, db)
		}
	})
}

func TestConflictMessageBoundsEscapedNames(t *testing.T) {
	for _, name := range []string{"model", strings.Repeat("中", 100), strings.Repeat("\x00", 100), strings.Repeat("\\", 100)} {
		got := ConflictMessage(name)
		if len(got) > 120 || !utf8.ValidString(got) {
			t.Fatalf("unsafe close reason %q", got)
		}
	}
}
