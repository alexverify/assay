package parse

import "testing"

func TestJSONCStripsCommentsAndTrailingCommas(t *testing.T) {
	input := []byte(`{
		// a line comment
		"mcpServers": {
			"a": { "command": "npx" }, /* block comment */
			"b": { "command": "node" }, // trailing comma below
		},
	}`)
	var got struct {
		MCPServers map[string]struct {
			Command string `json:"command"`
		} `json:"mcpServers"`
	}
	if err := JSONC(input, &got); err != nil {
		t.Fatalf("JSONC: %v", err)
	}
	if got.MCPServers["a"].Command != "npx" || got.MCPServers["b"].Command != "node" {
		t.Fatalf("parsed wrong: %+v", got.MCPServers)
	}
}

func TestJSONCPreservesCommentMarkersInsideStrings(t *testing.T) {
	input := []byte(`{ "url": "https://x/y//z", "note": "a /* not a comment */ b", "trail": "ends with ," }`)
	var got map[string]string
	if err := JSONC(input, &got); err != nil {
		t.Fatalf("JSONC: %v", err)
	}
	if got["url"] != "https://x/y//z" {
		t.Errorf("url mangled: %q", got["url"])
	}
	if got["note"] != "a /* not a comment */ b" {
		t.Errorf("note mangled: %q", got["note"])
	}
	if got["trail"] != "ends with ," {
		t.Errorf("trail mangled: %q", got["trail"])
	}
}

func TestJSONCHandlesEscapedQuotes(t *testing.T) {
	input := []byte(`{ "q": "he said \"hi\" // not a comment", }`)
	var got map[string]string
	if err := JSONC(input, &got); err != nil {
		t.Fatalf("JSONC: %v", err)
	}
	if got["q"] != `he said "hi" // not a comment` {
		t.Errorf("q mangled: %q", got["q"])
	}
}

func TestJSONCAcceptsPlainJSON(t *testing.T) {
	var got map[string]int
	if err := JSONC([]byte(`{"n": 1}`), &got); err != nil || got["n"] != 1 {
		t.Fatalf("plain JSON failed: %v %+v", err, got)
	}
}
