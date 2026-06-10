package discover

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alexverify/agentguard/internal/app/ports"
)

func TestDefaultDiscoversAcrossTools(t *testing.T) {
	dir := t.TempDir()
	// Claude Code + Cursor + Gemini configs side by side in one project.
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{"mcpServers":{"cc":{"command":"npx","args":["-y","cc@1.0.0"]}}}`)
	writeFile(t, filepath.Join(dir, ".cursor", "mcp.json"), `{"mcpServers":{"cur":{"url":"https://x/sse"}}}`)
	writeFile(t, filepath.Join(dir, ".gemini", "settings.json"), `{"mcpServers":{"gem":{"command":"npx","args":["-y","gem@2.0.0"]}}}`)

	got, err := Default().Discover(context.Background(), []ports.Scope{{Kind: "project", Path: dir}})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	tools := map[string]bool{}
	for _, a := range got {
		tools[a.Tool] = true
	}
	for _, want := range []string{"claude-code", "cursor", "gemini"} {
		if !tools[want] {
			t.Errorf("Default() did not discover tool %q; tools seen: %v", want, tools)
		}
	}
}
