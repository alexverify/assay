package sandbox

import (
	"strings"
	"testing"
)

func sampleProfile() Profile {
	return Profile{
		Workspace:  "/Users/dev/project",
		ProxyAddr:  "127.0.0.1:54321",
		DenyPaths:  []string{"/Users/dev/.ssh", "/Users/dev/.aws"},
		WritePaths: []string{"/private/var/folders/tmp"},
	}
}

func TestSeatbeltProfileShape(t *testing.T) {
	sb := seatbeltProfile(sampleProfile())

	mustContain(t, sb, "seatbelt", []string{
		"(version 1)",
		"(deny default)",
		"(allow file-read*)", // reads permissive
		`(allow file-write* (subpath "/Users/dev/project"))`,
		`(allow file-write* (subpath "/dev"))`,
		`(allow file-write* (subpath "/private/var/folders/tmp"))`,
		"(deny network*)",
		`(allow network* (remote tcp "localhost:54321"))`,
		`(deny file-read* (subpath "/Users/dev/.ssh"))`,
		`(deny file-write* (subpath "/Users/dev/.aws"))`,
	})

	// Process basics must be allowed or the server can't even start.
	mustContain(t, sb, "seatbelt", []string{"(allow process*)", "(allow mach*)"})

	// Ordering (Seatbelt is last-match-wins):
	if strings.Index(sb, "(deny network*)") > strings.Index(sb, `(allow network* (remote tcp "localhost:54321"))`) {
		t.Error("network deny must precede the proxy allow")
	}
	// The broad read allow must precede the secret denies, or they'd be undone.
	if strings.Index(sb, "(allow file-read*)") > strings.Index(sb, `(deny file-read* (subpath "/Users/dev/.ssh"))`) {
		t.Error("secret read-denies must come after the broad read allow")
	}
}

func TestSeatbeltProfileQuotesSafely(t *testing.T) {
	p := Profile{Workspace: `/tmp/has"quote`, ProxyAddr: "127.0.0.1:1"}
	sb := seatbeltProfile(p)
	// A naive embed would break the s-expression; the quote must be escaped.
	if strings.Contains(sb, `subpath "/tmp/has"quote"`) {
		t.Errorf("workspace path not escaped, profile is corruptible:\n%s", sb)
	}
}

func TestBwrapArgsShape(t *testing.T) {
	args := bwrapArgs(sampleProfile())
	joined := strings.Join(args, " ")

	mustContain(t, joined, "bwrap", []string{
		"--ro-bind / /",
		"--bind /Users/dev/project /Users/dev/project",
		"--tmpfs /Users/dev/.ssh",
		"--tmpfs /Users/dev/.aws",
		"--die-with-parent",
		"--unshare-pid",
	})

	// The workspace bind must come after the root ro-bind, or the read-only
	// root would shadow the writable workspace.
	if strings.Index(joined, "--ro-bind / /") > strings.Index(joined, "--bind /Users/dev/project") {
		t.Error("root ro-bind must precede the workspace bind")
	}
	// bwrap flags terminate before the command; callers append "-- argv".
	if strings.Contains(joined, "--") && !strings.HasSuffix(joined, "--die-with-parent") {
		// last flag should be a real flag, command is appended by Wrap
	}
}

func mustContain(t *testing.T, haystack, label string, needles []string) {
	t.Helper()
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			t.Errorf("%s profile missing %q\n---\n%s", label, n, haystack)
		}
	}
}
