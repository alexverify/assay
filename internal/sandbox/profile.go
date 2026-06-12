// Package sandbox confines a wrapped MCP server to an allowed workspace and
// forces its network through the egress proxy, using OS primitives (macOS
// Seatbelt, Linux bubblewrap). It is the slice that makes proxy routing
// mandatory rather than cooperative.
//
// Profile is pure configuration; the backends turn it into a command-line
// wrapper around the server's argv. The profile-to-args translation here is
// the heart of the package and is unit-tested without spawning anything.
package sandbox

import (
	"fmt"
	"strings"
)

// Profile describes the confinement to apply to one server.
type Profile struct {
	// Workspace is the single directory the server may read and write.
	Workspace string
	// ProxyAddr is the host:port of the egress proxy — the only network
	// endpoint the server may reach.
	ProxyAddr string
	// DenyPaths are credential/secret locations to block explicitly, on top
	// of the deny-by-default posture.
	DenyPaths []string
	// AllowReadPaths are extra read-only paths the runtime needs (system
	// libraries, interpreters) so the server can actually start.
	AllowReadPaths []string
}

// seatbeltProfile renders a macOS sandbox-exec (.sb) profile. Seatbelt uses
// last-match-wins, so the order is: deny default → broad allows the runtime
// needs → workspace read/write → network deny then the single proxy allow →
// explicit secret-path denies last so nothing re-opens them.
func seatbeltProfile(p Profile) string {
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	w("(version 1)")
	w("(deny default)")

	// Process and system basics — without these even an interpreter won't load.
	w("(allow process-exec)")
	w("(allow process-fork)")
	w("(allow sysctl-read)")
	w("(allow mach-lookup)")
	w("(allow signal (target self))")
	w("(allow file-read-metadata)")

	// Read-only system paths the runtime needs.
	for _, path := range append([]string{"/usr", "/System", "/Library", "/bin", "/sbin", "/private/var", "/dev", "/etc"}, p.AllowReadPaths...) {
		w(fmt.Sprintf("(allow file-read* (subpath %s))", quote(path)))
	}

	// The workspace: full read/write.
	if p.Workspace != "" {
		w(fmt.Sprintf("(allow file-read* (subpath %s))", quote(p.Workspace)))
		w(fmt.Sprintf("(allow file-write* (subpath %s))", quote(p.Workspace)))
	}

	// Network: deny everything, then re-allow only the loopback proxy.
	w("(deny network*)")
	if p.ProxyAddr != "" {
		w(fmt.Sprintf("(allow network* (remote tcp %s))", quote(p.ProxyAddr)))
	}

	// Sensitive paths denied last so no earlier allow exposes them.
	for _, path := range p.DenyPaths {
		w(fmt.Sprintf("(deny file-read* (subpath %s))", quote(path)))
		w(fmt.Sprintf("(deny file-write* (subpath %s))", quote(path)))
	}
	return b.String()
}

// bwrapArgs renders bubblewrap flags: a read-only root, a read-write
// workspace bound over it, secret paths masked with empty tmpfs, and the
// process tied to its parent. Network is left to the injected HTTP(S)_PROXY
// plus these binds (no veth in this slice); callers append "-- argv".
func bwrapArgs(p Profile) []string {
	args := []string{
		"--ro-bind", "/", "/",
		"--dev", "/dev",
		"--proc", "/proc",
		"--unshare-pid",
		"--die-with-parent",
	}
	if p.Workspace != "" {
		args = append(args, "--bind", p.Workspace, p.Workspace)
	}
	for _, path := range p.DenyPaths {
		args = append(args, "--tmpfs", path)
	}
	return args
}

// quote renders a string as a Scheme-style double-quoted literal for the
// Seatbelt profile, escaping embedded quotes and backslashes.
func quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + r.Replace(s) + `"`
}
