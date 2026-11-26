package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events" || r.URL.Query().Get("owners") != "alice,org" || r.URL.Query().Get("staff") != "1" || r.Header.Get("Authorization") != "Bearer test" || r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected request: %v %v", r.URL, r.Header)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("denied"))
	}))
	defer server.Close()
	body, status, err := NewClient(server.Client(), server.URL).ListEvents("test", "alice,org", true)
	if err != nil || status != http.StatusUnauthorized || body != "denied" {
		t.Fatalf("got %q %d %v", body, status, err)
	}
}
