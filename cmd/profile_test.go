package cmd

import "testing"

func TestProfileUsage(t *testing.T) {
	total := 8.0
	available := 2.0
	if got, want := profileUsage(&total, &available), "6.00/8.00GB (75%)"; got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
	if got := profileUsage(nil, &available); got != "n/a" {
		t.Fatalf("want n/a for unavailable total, got %q", got)
	}
}
