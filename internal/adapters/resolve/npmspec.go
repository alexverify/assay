package resolve

import (
	"bytes"
	"encoding/json"
	"strings"
)

// parseNPMSpec splits an npm spec into name, version, and whether the version
// is an exact pin. Handles scoped packages (e.g. @scope/name@1.2.3).
func parseNPMSpec(ref string) (name, version string, pinned bool) {
	if ref == "" {
		return "", "", false
	}
	at := strings.LastIndex(ref, "@")
	if at <= 0 { // no version, or only the leading @ of a scope
		return ref, "", false
	}
	name = ref[:at]
	version = ref[at+1:]
	return name, version, isExactVersion(version)
}

// isExactVersion reports whether v is a single concrete version (no range
// operators, tags, or wildcards).
func isExactVersion(v string) bool {
	if v == "" || v == "latest" || v == "*" || v == "next" {
		return false
	}
	if strings.ContainsAny(v, "^~ ><=|*x") {
		return false
	}
	return v[0] >= '0' && v[0] <= '9'
}

// parseNPMStringOutput extracts a string from `npm view ... --json` output,
// which may be a JSON string or (when multiple versions match) a JSON array;
// in the latter case the last element is returned.
func parseNPMStringOutput(b []byte) string {
	b = bytes.TrimSpace(b)
	var s string
	if json.Unmarshal(b, &s) == nil {
		return s
	}
	var arr []string
	if json.Unmarshal(b, &arr) == nil && len(arr) > 0 {
		return arr[len(arr)-1]
	}
	return strings.Trim(string(b), `"`)
}
