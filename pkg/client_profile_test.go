package pkg

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProfiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/snapshots" {
			t.Fatalf("want snapshots path, got %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("owners"); got != "alexellis" {
			t.Fatalf("want owner alexellis, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "50" {
			t.Fatalf("want limit 50, got %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("want bearer token, got %q", got)
		}
		io.WriteString(w, "[]")
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	body, status, err := client.ListProfiles("token", "alexellis", 50, false)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK || body != "[]" {
		t.Fatalf("want 200 and empty list, got %d and %q", status, body)
	}
}

func TestGetProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/snapshot" {
			t.Fatalf("want snapshot path, got %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("owner"); got != "openfaasltd" {
			t.Fatalf("want owner openfaasltd, got %q", got)
		}
		if got := r.URL.Query().Get("job"); got != "123" {
			t.Fatalf("want job 123, got %q", got)
		}
		io.WriteString(w, `{}`)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	_, status, err := client.GetProfile("token", "openfaasltd", "123", false)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("want 200, got %d", status)
	}
}
