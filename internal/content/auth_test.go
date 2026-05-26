package content

import "testing"

func TestIsPublisherCaseInsensitive(t *testing.T) {
	allowed := []string{"Publisher@Client.com", " editor@example.com "}

	if !IsPublisher("publisher@client.com", allowed) {
		t.Fatal("expected case-insensitive match for publisher email")
	}
	if !IsPublisher("  editor@example.com", allowed) {
		t.Fatal("expected trimmed case-insensitive match")
	}
	if IsPublisher("other@example.com", allowed) {
		t.Fatal("unexpected match for non-publisher email")
	}
}
