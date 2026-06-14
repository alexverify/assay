// Package notify posts short digests to a chatops webhook. It is deliberately
// tiny: the Slack "incoming webhook" JSON shape ({"text": ...}) is accepted by
// Slack, Mattermost, Discord (via compat), and most generic receivers, so one
// code path covers the common cases without per-vendor adapters.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Sender posts notifications. The zero value is usable (falls back to
// http.DefaultClient); New is provided for symmetry with the other adapters.
type Sender struct {
	Client *http.Client
}

// New returns a Sender backed by the default HTTP client.
func New() Sender { return Sender{Client: http.DefaultClient} }

// Post sends text to the webhook URL as a {"text": ...} JSON body. It returns
// an error on transport failure or any non-2xx response.
func (s Sender) Post(ctx context.Context, url, text string) error {
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook %s returned %s", url, resp.Status)
	}
	return nil
}
