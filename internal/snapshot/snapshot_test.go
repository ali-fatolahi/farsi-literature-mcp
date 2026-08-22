package snapshot

import "testing"

func TestValidateManifest(t *testing.T) {
	valid := Manifest{
		SchemaVersion: 1, GeneratedAtUTC: "2026-08-16T12:20:22Z",
		PoetsCount: 1, PoemsCount: 1, IDIndexShardSize: 2000,
		URLTemplates: URLTemplates{Poet: "poet", Category: "category", Poem: "poem"},
		Poets:        []ManifestPoet{{ID: 2, FullURL: "/hafez"}},
	}
	if err := validateManifest(valid); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	valid.SchemaVersion = 2
	if err := validateManifest(valid); err == nil {
		t.Fatal("unsupported schema version accepted")
	}
}

func TestContentPathRejectsTraversal(t *testing.T) {
	if _, err := contentPath("/../escape", ".json"); err == nil {
		t.Fatal("path traversal accepted")
	}
}

func TestContentPath(t *testing.T) {
	tests := []struct {
		url, suffix, want string
	}{
		{"/hafez", "poet.json", "poets/hafez/poet.json"},
		{"/hafez/ghazal", "_cat.json", "poets/hafez/ghazal/_cat.json"},
		{"/hafez/ghazal/sh1", ".json", "poets/hafez/ghazal/sh1.json"},
	}
	for _, test := range tests {
		got, err := contentPath(test.url, test.suffix)
		if err != nil {
			t.Fatalf("contentPath(%q, %q): %v", test.url, test.suffix, err)
		}
		if got != test.want {
			t.Errorf("contentPath(%q, %q) = %q, want %q", test.url, test.suffix, got, test.want)
		}
	}
}
