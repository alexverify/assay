package reputation

import (
	"testing"
	"time"
)

func TestLookupByHashIsCaseInsensitive(t *testing.T) {
	src := Source{
		"abc123": {Hash: "abc123", Trusters: 42, FirstSeen: time.Unix(1000, 0).UTC()},
	}
	got, ok := src.Lookup("ABC123")
	if !ok {
		t.Fatalf("expected a hit for a differently-cased hash")
	}
	if got.Trusters != 42 {
		t.Errorf("trusters = %d, want 42", got.Trusters)
	}
}

func TestLookupMissIsUnknownNotNegative(t *testing.T) {
	// An absent hash returns false — never a claim that the artifact is bad.
	if _, ok := (Source{}).Lookup("deadbeef"); ok {
		t.Errorf("an empty corpus must never report a hit")
	}
	if _, ok := Source(nil).Lookup("x"); ok {
		t.Errorf("a nil corpus must be a silent no-op")
	}
}

func TestGradeTiers(t *testing.T) {
	cases := []struct {
		trusters int
		want     Grade
	}{
		{0, GradeUnknown},
		{1, GradeEmerging},
		{9, GradeEmerging},
		{10, GradeEstablished},
		{500, GradeEstablished},
	}
	for _, c := range cases {
		if got := (Signal{Trusters: c.trusters}).Grade(); got != c.want {
			t.Errorf("Grade(trusters=%d) = %q, want %q", c.trusters, got, c.want)
		}
	}
}

func TestLookupEmptyHashIsNoOp(t *testing.T) {
	src := Source{"": {Trusters: 5}}
	if _, ok := src.Lookup(""); ok {
		t.Errorf("an empty hash must never resolve, even if the corpus has a blank key")
	}
}
