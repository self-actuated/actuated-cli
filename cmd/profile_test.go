package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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

func TestProfileDetailsFormatsRawMemoryBytes(t *testing.T) {
	total := 12_260_000_000.0
	available := 9_920_000_000.0
	var output bytes.Buffer

	printProfileDetails(&output, ProfileSnapshot{
		TotalMemoryBytes:        &total,
		MinAvailableMemoryBytes: &available,
	})

	for _, want := range []string{"12.26GB", "9.92GB"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("want output to contain %q, got:\n%s", want, output.String())
		}
	}
}
