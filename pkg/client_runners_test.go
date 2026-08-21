package pkg

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRunnersRequestsVerboseOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/v1/runners"; got != want {
			t.Fatalf("want path %q, got %q", want, got)
		}
		if got, want := r.URL.Query().Get("owner"), "alexellis"; got != want {
			t.Fatalf("want owner %q, got %q", want, got)
		}
		if got, want := r.URL.Query().Get("verbose"), "1"; got != want {
			t.Fatalf("want verbose %q, got %q", want, got)
		}
		if got := r.Header.Get("Accept"); got != "" {
			t.Fatalf("want plain-text Accept header, got %q", got)
		}
		io.WriteString(w, "table")
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	body, status, err := client.ListRunners("token", "alexellis", false, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK || body != "table" {
		t.Fatalf("want 200 and table, got %d and %q", status, body)
	}
}
