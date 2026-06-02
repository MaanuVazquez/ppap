package input

import (
	"math"
	"testing"
)

func TestPenEventValidateAcceptsNormalizedEvent(t *testing.T) {
	event := PenEvent{
		Type:     PenEventDown,
		X:        0.5,
		Y:        0.25,
		Pressure: 0.75,
	}

	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid event: %v", err)
	}
}

func TestPenEventValidateRejectsInvalidValues(t *testing.T) {
	tests := []PenEvent{
		{Type: PenEventType("tap"), X: 0.5, Y: 0.5, Pressure: 0.5},
		{Type: PenEventDown, X: -0.1, Y: 0.5, Pressure: 0.5},
		{Type: PenEventDown, X: 0.5, Y: 1.1, Pressure: 0.5},
		{Type: PenEventDown, X: 0.5, Y: 0.5, Pressure: math.NaN()},
		{Type: PenEventDown, X: 0.5, Y: 0.5, Pressure: math.Inf(1)},
	}

	for _, event := range tests {
		if err := event.Validate(); err == nil {
			t.Fatalf("expected invalid event: %+v", event)
		}
	}
}
