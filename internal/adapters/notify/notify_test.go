package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostSendsTextPayload(t *testing.T) {
	var got map[string]string
	var ct string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := New().Post(context.Background(), srv.URL, "1 drifted, 0 new"); err != nil {
		t.Fatal(err)
	}
	if got["text"] != "1 drifted, 0 new" {
		t.Errorf("payload text = %q", got["text"])
	}
	if ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func TestPostErrorsOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if err := New().Post(context.Background(), srv.URL, "x"); err == nil {
		t.Fatal("expected an error on a 500 response")
	}
}
