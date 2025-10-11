package reverse_proxy

import "testing"

func TestTruncateBodyForAudit(t *testing.T) {
	body := []byte("abcdefghijklmnopqrstuvwxyz")
	got := truncateBodyForAudit(body, 5, 3)
	want := "abcde...[omitted 18 bytes]...xyz"
	if string(got) != want {
		t.Fatalf("unexpected truncation result: %q", string(got))
	}
}
