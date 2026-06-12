package policy

import "path"

// MCPPolicy is the runtime gate the shim enforces on MCP tool calls. Keyed by
// server name; the "*" entry applies to every server.
type MCPPolicy struct {
	Servers map[string]ToolRule `json:"servers,omitempty"`
}

// ToolRule constrains which tools a server may be asked to run. Patterns use
// path.Match syntax ("delete_*", "get_?"); a plain name matches exactly.
type ToolRule struct {
	// AllowTools, when non-empty, is exhaustive: only matching tools may run.
	AllowTools []string `json:"allowTools,omitempty"`
	// DenyTools always blocks matching tools, even when allowlisted.
	DenyTools []string `json:"denyTools,omitempty"`
}

// Decision is the outcome of DecideTool, with the matched rule for the audit
// trail.
type Decision struct {
	Allowed bool
	Reason  string // the matching pattern (denials) or unmatched allowlist
}

// DecideTool applies the MCP rules to one tools/call. Semantics, in order:
//
//  1. A deny match — on the server's entry or the "*" entry — always blocks.
//  2. If the server's entry (falling back to "*") has an allowlist, the tool
//     must match it.
//  3. Otherwise the call is allowed: absent rules must never break a session.
func (p Policy) DecideTool(server, tool string) Decision {
	specific, hasSpecific := p.MCP.Servers[server]
	star := p.MCP.Servers["*"]

	if pat := matchAny(specific.DenyTools, tool); pat != "" {
		return Decision{Allowed: false, Reason: "denied by " + server + " denyTools " + pat}
	}
	if pat := matchAny(star.DenyTools, tool); pat != "" {
		return Decision{Allowed: false, Reason: "denied by * denyTools " + pat}
	}

	allow := specific.AllowTools
	scope := server
	if !hasSpecific || len(allow) == 0 {
		allow = star.AllowTools
		scope = "*"
	}
	if len(allow) > 0 && matchAny(allow, tool) == "" {
		return Decision{Allowed: false, Reason: "not in " + scope + " allowTools"}
	}
	return Decision{Allowed: true}
}

// matchAny returns the first pattern matching the tool, or "".
func matchAny(patterns []string, tool string) string {
	for _, pat := range patterns {
		if ok, err := path.Match(pat, tool); err == nil && ok {
			return pat
		}
	}
	return ""
}
