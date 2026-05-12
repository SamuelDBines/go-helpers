package env

import (
	"testing"
	"time"
)

func TestIntDefault(t *testing.T) {
	t.Setenv("ENV_TEST_INT", "")
	if got := IntDefault("ENV_TEST_INT", 7); got != 7 {
		t.Fatalf("empty: got %d", got)
	}
	t.Setenv("ENV_TEST_INT", "42")
	if got := IntDefault("ENV_TEST_INT", 7); got != 42 {
		t.Fatalf("valid: got %d", got)
	}
	t.Setenv("ENV_TEST_INT", "nope")
	if got := IntDefault("ENV_TEST_INT", 7); got != 7 {
		t.Fatalf("invalid: got %d", got)
	}
}

func TestDuration(t *testing.T) {
	t.Setenv("ENV_TEST_DUR", "")
	if got := Duration("ENV_TEST_DUR", time.Minute); got != time.Minute {
		t.Fatalf("empty: got %v", got)
	}
	t.Setenv("ENV_TEST_DUR", "30s")
	if got := Duration("ENV_TEST_DUR", time.Minute); got != 30*time.Second {
		t.Fatalf("valid: got %v", got)
	}
	t.Setenv("ENV_TEST_DUR", "not-a-duration")
	if got := Duration("ENV_TEST_DUR", time.Minute); got != time.Minute {
		t.Fatalf("invalid: got %v", got)
	}
}

func TestCommaList(t *testing.T) {
	if got := CommaList("ENV_TEST_CSV_UNSET"); got != nil {
		t.Fatalf("unset: %v", got)
	}
	t.Setenv("ENV_TEST_CSV", " a , , b ")
	got := CommaList("ENV_TEST_CSV")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}
