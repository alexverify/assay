package controlplane

import (
	"crypto/subtle"
	"strings"
)

// Auth resolves a bearer token to the org it is scoped to. A machine token is
// issued by an org admin and identifies which org's fleet a submission joins
// and which org's report a read returns — the row-level isolation boundary.
type Auth interface {
	Org(token string) (string, bool)
}

// StaticAuth is a fixed token → org map, configured at server start (the v1
// auth). OIDC for humans and admin-issued token management come in later slices.
type StaticAuth map[string]string

// Org returns the org a token is scoped to. The comparison is constant-time and
// always scans every configured token, so a caller cannot learn which token (if
// any) it was close to from timing. An empty token never matches.
func (a StaticAuth) Org(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	var org string
	var ok bool
	for t, o := range a {
		if subtle.ConstantTimeCompare([]byte(t), []byte(token)) == 1 {
			org, ok = o, true
		}
	}
	return org, ok
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}
