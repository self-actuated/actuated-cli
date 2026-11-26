package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v76/github"
)

func TestEventOwnersPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			fmt.Fprint(w, `{"login":"alice"}`)
		case "/user/memberships/orgs":
			if r.URL.Query().Get("page") == "2" {
				fmt.Fprint(w, `[{"state":"active","organization":{"login":"beta"}}]`)
				return
			}
			w.Header().Set("Link", fmt.Sprintf(`<%s?page=2>; rel="next"`, "http://"+r.Host+r.URL.Path))
			fmt.Fprint(w, `[{"state":"active","organization":{"login":"alpha"}},{"state":"pending","organization":{"login":"ignored"}}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := github.NewClient(server.Client())
	client.BaseURL, _ = url.Parse(server.URL + "/")
	command := makeEvents()
	command.SetContext(context.Background())
	owners, err := eventOwners(client, command)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(owners) != "[alice alpha beta]" {
		t.Fatalf("owners: %v", owners)
	}
}

func TestFilterEvents(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	events := []actuatedEvent{{Message: "old", CreatedAt: now.Add(-2 * time.Hour)}, {Message: "new", CreatedAt: now}, {Message: "boundary", CreatedAt: now.Add(-time.Hour)}}
	got := filterEvents(events, time.Hour, now)
	if len(got) != 2 || got[0].Message != "new" || got[1].Message != "boundary" {
		t.Fatalf("events: %+v", got)
	}
	empty, err := json.Marshal(filterEvents(nil, 0, now))
	if err != nil || string(empty) != "[]" {
		t.Fatalf("empty JSON: %s %v", empty, err)
	}
}
