package contextcard

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/quick"
)

func TestRenderTemplateParseRoundTripProperty(t *testing.T) {
	property := func(input quickCardInput) bool {
		content := renderTemplate(input.Slug, input.Scope, input.Tags)
		card, err := Parse(cardPath(input.Slug), []byte(content))
		if err != nil {
			t.Logf("Parse(renderTemplate(...)) returned error: %v\n%s", err, content)
			return false
		}

		if card.Slug != input.Slug || card.Path != cardPath(input.Slug) {
			t.Logf("card identity = (%q, %q), want (%q, %q)", card.Slug, card.Path, input.Slug, cardPath(input.Slug))
			return false
		}
		if !slices.Equal(card.Scope, input.Scope) {
			t.Logf("scope = %#v, want %#v", card.Scope, input.Scope)
			return false
		}
		if !slices.Equal(card.Tags, input.Tags) {
			t.Logf("tags = %#v, want %#v", card.Tags, input.Tags)
			return false
		}
		if card.Body != "# "+titleFromSlug(input.Slug)+"\n\nWrite the repo context here." {
			t.Logf("body = %q", card.Body)
			return false
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

type quickCardInput struct {
	Slug  string
	Scope []string
	Tags  []string
}

func (quickCardInput) Generate(r *rand.Rand, _ int) reflect.Value {
	input := quickCardInput{
		Slug:  generatedCardSlug(r),
		Scope: make([]string, 1+r.Intn(4)),
		Tags:  make([]string, r.Intn(4)),
	}
	for i := range input.Scope {
		input.Scope[i] = generatedCardScope(r)
	}
	for i := range input.Tags {
		input.Tags[i] = generatedCardTag(r)
	}
	return reflect.ValueOf(input)
}

func generatedCardScope(r *rand.Rand) string {
	prefixes := []string{"cmd", "docs", "internal", ".garden/context"}
	return prefixes[r.Intn(len(prefixes))] + "/" + generatedCardSlug(r) + "/**"
}

func generatedCardTag(r *rand.Rand) string {
	return generatedCardSlug(r)
}

func generatedCardSlug(r *rand.Rand) string {
	parts := make([]string, 1+r.Intn(3))
	for i := range parts {
		parts[i] = generatedCardWord(r)
	}
	return strings.Join(parts, "-")
}

func generatedCardWord(r *rand.Rand) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	for range 1 + r.Intn(8) {
		b.WriteByte(alphabet[r.Intn(len(alphabet))])
	}
	return b.String()
}
