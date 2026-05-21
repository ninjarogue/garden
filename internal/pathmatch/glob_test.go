package pathmatch

import "testing"

func TestAnyMatchesRepoRelativeDoubleStarGlobs(t *testing.T) {
	matched, err := Any([]string{"src/routes/**"}, "./src/routes/api/users.ts")
	if err != nil {
		t.Fatalf("Any returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected scope to match path")
	}
}

func TestAnyDoesNotMatchDifferentScope(t *testing.T) {
	matched, err := Any([]string{"docs/**"}, "src/routes/api/users.ts")
	if err != nil {
		t.Fatalf("Any returned error: %v", err)
	}
	if matched {
		t.Fatal("expected docs scope not to match route path")
	}
}
