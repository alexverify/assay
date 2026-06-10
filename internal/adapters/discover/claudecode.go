package discover

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alexverify/agentguard/internal/adapters/parse"
	"github.com/alexverify/agentguard/internal/app/ports"
	"github.com/alexverify/agentguard/internal/domain/artifact"
)

// ClaudeCode discovers Claude Code MCP servers and skills.
type ClaudeCode struct {
	home string
}

// NewClaudeCode constructs the discoverer, resolving the user's home directory
// for global-scope lookups.
func NewClaudeCode() *ClaudeCode {
	home, _ := os.UserHomeDir()
	return &ClaudeCode{home: home}
}

// Tool returns the canonical tool id.
func (c *ClaudeCode) Tool() string { return "claude-code" }

// Discover satisfies ports.Discoverer.
func (c *ClaudeCode) Discover(_ context.Context, scopes []ports.Scope) ([]artifact.Artifact, error) {
	var out []artifact.Artifact
	for _, sc := range scopes {
		switch sc.Kind {
		case "project":
			out = append(out, mcpServersFromConfig(c.Tool(), filepath.Join(sc.Path, ".mcp.json"), sc.String(), parse.JSON)...)
			out = append(out, skillsFromDir(c.Tool(), filepath.Join(sc.Path, ".claude", "skills"), sc.String())...)
		case "global":
			if c.home != "" {
				out = append(out, mcpServersFromConfig(c.Tool(), filepath.Join(c.home, ".claude.json"), "global", parse.JSON)...)
				out = append(out, skillsFromDir(c.Tool(), filepath.Join(c.home, ".claude", "skills"), "global")...)
			}
		}
	}
	return out, nil
}
