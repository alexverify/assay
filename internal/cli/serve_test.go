package cli_test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexverify/assay/internal/cli"
	"github.com/alexverify/assay/internal/controlplane"
	"github.com/alexverify/assay/internal/domain/policy"
)

func TestFleetPushThenShowRemote(t *testing.T) {
	// Stand up a real control-plane server, push this fixture project's snapshot
	// to it, then read the aggregated fleet back through the same CLI.
	srv := httptest.NewServer(controlplane.NewServer(
		controlplane.NewService(controlplane.NewMemStore(), nil),
		controlplane.StaticAuth{"tok": "acme"},
	))
	t.Cleanup(srv.Close)

	dir, lock := fixtureProject(t)
	ctx := context.Background()

	push, out, errBuf := newApp()
	code := push.Execute(ctx, []string{
		"fleet", "push", "--path", dir, "--lockfile", lock,
		"--owner", "alice", "--server", srv.URL, "--token", "tok",
	})
	if code != cli.ExitOK {
		t.Fatalf("push exit = %d, stderr=%s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "pushed fleet snapshot") {
		t.Errorf("push should confirm: %s", out.String())
	}

	show, sout, _ := newApp()
	code = show.Execute(ctx, []string{"fleet", "--server", srv.URL, "--token", "tok"})
	if code != cli.ExitOK {
		t.Fatalf("remote show exit = %d", code)
	}
	if !strings.Contains(sout.String(), "1 machines") {
		t.Errorf("remote fleet should report alice's machine:\n%s", sout.String())
	}
}

func TestFleetPushRequiresServer(t *testing.T) {
	dir, lock := fixtureProject(t)
	app, _, _ := newApp()
	code := app.Execute(context.Background(), []string{"fleet", "push", "--path", dir, "--lockfile", lock})
	if code != cli.ExitUsage {
		t.Errorf("push without a server should be a usage error, got %d", code)
	}
}

func TestServeRequiresTokens(t *testing.T) {
	app, _, errBuf := newApp()
	code := app.Execute(context.Background(), []string{"serve", "--addr", "127.0.0.1:0"})
	if code != cli.ExitUsage {
		t.Fatalf("serve without --tokens should be a usage error, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "tokens") {
		t.Errorf("error should mention tokens: %s", errBuf.String())
	}
}

func TestFleetVerifyPullsServerPolicy(t *testing.T) {
	// The server holds a strict blast-radius policy; a local fleet dir with a
	// drift on two machines must fail the gate when verify pulls that policy,
	// even though no local policy file exists.
	cfg := controlplane.NewMemConfig()
	cfg.SetPolicy("acme", policy.Policy{Fleet: policy.FleetPolicy{MaxBlastRadius: 1}})
	srv := httptest.NewServer(controlplane.NewServer(
		controlplane.NewService(controlplane.NewMemStore(), cfg),
		controlplane.StaticAuth{"tok": "acme"},
	))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	for _, o := range []string{"alice", "bob"} {
		body := `{"owner":"` + o + `","artifacts":[{"id":"x","name":"feed","kind":"skill","hash":"h-` + o + `","drift":"drifted","verdict":"review"}]}`
		if err := os.WriteFile(filepath.Join(dir, o+".json"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	app, _, _ := newApp()
	// Point --policy at a nonexistent file so only the server policy can gate.
	code := app.Execute(context.Background(), []string{
		"fleet", "verify", "--dir", dir,
		"--policy", filepath.Join(dir, "none.json"),
		"--server", srv.URL, "--token", "tok",
	})
	if code != cli.ExitDrift {
		t.Fatalf("server policy should fail the gate (1), got %d", code)
	}
}

func TestServeRejectsEmptyTokensFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	app, _, _ := newApp()
	code := app.Execute(context.Background(), []string{"serve", "--addr", "127.0.0.1:0", "--tokens", path})
	if code != cli.ExitUsage {
		t.Errorf("an empty tokens file should be rejected, got %d", code)
	}
}
