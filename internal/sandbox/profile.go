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

// Profile describes the confinement to apply to one server. The posture is
// permissive reads (so the runtime and its libraries load), locked-down
// writes (only the workspace and scratch dirs), locked-down network (only the
// egress proxy), and explicit denies for credential paths — matching the
// spec's "deny ~/.ssh, ~/.aws, …" model.
//
// Paths must be symlink-resolved by the caller: Seatbelt and bwrap match the
// real path, and on macOS /tmp and /var are symlinks into /private.
type Profile struct {
	// Workspace is the directory the server may read and write.
	Workspace string
	// ProxyAddr is the host:port of the egress proxy — the only network
	// endpoint the server may reach.
	ProxyAddr string
	// DenyPaths are credential/secret locations blocked for read and write,
	// overriding the permissive read posture.
	DenyPaths []string
	// WritePaths are extra writable directories the runtime needs (e.g. the
	// per-user temp dir). The workspace and /dev are always writable.
	WritePaths []string
}

// seatbeltProfile renders a macOS sandbox-exec (.sb) profile. Seatbelt is
// last-match-wins, so order matters: deny default → process/system allows →
// broad read allow → write allows → network deny then proxy allow → secret
// denies last so nothing re-exposes them.
func seatbeltProfile(p Profile) string {
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	w("(version 1)")
	w("(deny default)")

	// Process and IPC basics — without these even an interpreter won't load.
	w("(allow process*)")
	w("(allow mach*)")
	w("(allow sysctl*)")
	w("(allow signal (target self))")
	w("(allow ipc*)")

	// Reads are permissive; specific secret paths are denied below.
	w("(allow file-read*)")

	// Writes: workspace, scratch dirs, and devices only.
	if p.Workspace != "" {
		w(fmt.Sprintf("(allow file-write* (subpath %s))", quote(p.Workspace)))
	}
	for _, path := range append([]string{"/dev"}, p.WritePaths...) {
		w(fmt.Sprintf("(allow file-write* (subpath %s))", quote(path)))
	}

	// Network: deny everything, then re-allow only the loopback proxy.
	// Seatbelt requires the host to be "localhost" or "*" (not a literal IP);
	// the proxy listens on loopback, so localhost reaches it.
	w("(deny network*)")
	if port := proxyPort(p.ProxyAddr); port != "" {
		w(fmt.Sprintf("(allow network* (remote tcp %s))", quote("localhost:"+port)))
	}

	// Sensitive paths denied last so no earlier allow exposes them.
	for _, path := range p.DenyPaths {
		w(fmt.Sprintf("(deny file-read* (subpath %s))", quote(path)))
		w(fmt.Sprintf("(deny file-write* (subpath %s))", quote(path)))
	}
	return b.String()
}

// bwrapArgs renders bubblewrap flags. A read-only root bind makes everything
// readable but unwritable (matching the permissive-read posture); the
// workspace and scratch dirs are bound read-write over it; secret paths are
// masked with an empty tmpfs (denying read and write); the process is tied to
// its parent. Network confinement is left to the injected HTTP(S)_PROXY in
// this slice (no veth). Callers append "-- argv".
func bwrapArgs(p Profile) []string {
	args := []string{
		"--ro-bind", "/", "/",
		"--dev", "/dev",
		"--proc", "/proc",
		"--tmpfs", "/tmp",
		"--unshare-pid",
		"--die-with-parent",
	}
	if p.Workspace != "" {
		args = append(args, "--bind", p.Workspace, p.Workspace)
	}
	for _, path := range p.WritePaths {
		args = append(args, "--bind", path, path)
	}
	// Mask secrets last so they win over any bind above.
	for _, path := range p.DenyPaths {
		args = append(args, "--tmpfs", path)
	}
	return args
}

// proxyPort returns the port from a host:port address, or "" if absent.
func proxyPort(addr string) string {
	if i := strings.LastIndexByte(addr, ':'); i >= 0 {
		return addr[i+1:]
	}
	return ""
}

// quote renders a string as a Scheme-style double-quoted literal for the
// Seatbelt profile, escaping embedded quotes and backslashes.
func quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + r.Replace(s) + `"`
}
