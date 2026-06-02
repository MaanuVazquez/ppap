package display

import "testing"

func TestBoundsEdges(t *testing.T) {
	bounds := Bounds{Left: -1920, Top: 100, Width: 3840, Height: 2160}

	if got := bounds.Right(); got != 1920 {
		t.Fatalf("expected right edge 1920, got %d", got)
	}
	if got := bounds.Bottom(); got != 2260 {
		t.Fatalf("expected bottom edge 2260, got %d", got)
	}
}
