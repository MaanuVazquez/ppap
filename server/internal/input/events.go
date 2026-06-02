package input

import (
	"fmt"
	"math"
)

type PenEventType string

const (
	PenEventDown PenEventType = "down"
	PenEventMove PenEventType = "move"
	PenEventUp   PenEventType = "up"
)

type PenEvent struct {
	Type      PenEventType `json:"type"`
	X         float64      `json:"x"`
	Y         float64      `json:"y"`
	Pressure  float64      `json:"pressure"`
	PointerID int          `json:"pointerId"`
	TiltX     float64      `json:"tiltX"`
	TiltY     float64      `json:"tiltY"`
	Twist     float64      `json:"twist"`
}

func (event PenEvent) Validate() error {
	switch event.Type {
	case PenEventDown, PenEventMove, PenEventUp:
	default:
		return fmt.Errorf("unsupported pen event type %q", event.Type)
	}

	if !isNormalized(event.X) {
		return fmt.Errorf("x must be normalized between 0 and 1")
	}
	if !isNormalized(event.Y) {
		return fmt.Errorf("y must be normalized between 0 and 1")
	}
	if !isNormalized(event.Pressure) {
		return fmt.Errorf("pressure must be normalized between 0 and 1")
	}

	return nil
}

func isNormalized(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
