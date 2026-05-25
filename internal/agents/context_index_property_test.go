package agents

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/quick"
)

func TestRenderIndexIsIndependentOfCardOrderProperty(t *testing.T) {
	property := func(cards quickIndexCards) bool {
		original := append([]IndexCard(nil), []IndexCard(cards)...)
		reversed := append([]IndexCard(nil), original...)
		slices.Reverse(reversed)

		got, err := RenderIndex(original)
		if err != nil {
			t.Logf("RenderIndex(original) returned error: %v", err)
			return false
		}
		want, err := RenderIndex(reversed)
		if err != nil {
			t.Logf("RenderIndex(reversed) returned error: %v", err)
			return false
		}
		if got != want {
			t.Logf("RenderIndex differs after reversing input\noriginal:\n%s\nreversed:\n%s", got, want)
			return false
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

type quickIndexCards []IndexCard

func (quickIndexCards) Generate(r *rand.Rand, _ int) reflect.Value {
	count := r.Intn(12)
	cards := make([]IndexCard, count)
	for i := range cards {
		scopes := make([]string, r.Intn(4))
		for j := range scopes {
			scopes[j] = generatedScope(r)
		}
		cards[i] = IndexCard{
			Path:  ".garden/context/" + generatedSlug(r) + ".md",
			Scope: scopes,
		}
	}
	return reflect.ValueOf(quickIndexCards(cards))
}

func generatedScope(r *rand.Rand) string {
	prefixes := []string{"cmd", "docs", "internal", ".garden/context"}
	return prefixes[r.Intn(len(prefixes))] + "/" + generatedSlug(r) + "/**"
}

func generatedSlug(r *rand.Rand) string {
	parts := make([]string, 1+r.Intn(3))
	for i := range parts {
		parts[i] = generatedWord(r)
	}
	return strings.Join(parts, "-")
}

func generatedWord(r *rand.Rand) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	for range 1 + r.Intn(8) {
		b.WriteByte(alphabet[r.Intn(len(alphabet))])
	}
	return b.String()
}
