// Package parse normalizes the heterogeneous config formats used by AI coding
// tools into Go values. JSON and JSONC (Claude/Cursor/VS Code-style, with
// comments and trailing commas) are supported. TOML (Codex) is a documented
// seam: it returns ErrUnsupportedFormat until a parser is wired, so discovery
// can note and skip rather than mis-parse.
package parse

import (
	"encoding/json"
	"errors"
)

// ErrUnsupportedFormat indicates a config format whose parser is not yet wired.
var ErrUnsupportedFormat = errors.New("config format not yet supported")

// JSON unmarshals strict JSON into v.
func JSON(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

// TOML parses TOML (e.g. Codex config.toml). Not yet implemented; see package
// docs. Adding it is an isolated change behind this seam.
func TOML(_ []byte, _ any) error {
	return ErrUnsupportedFormat
}
