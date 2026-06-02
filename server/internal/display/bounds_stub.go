//go:build !windows

package display

func VirtualBounds() Bounds {
	return Bounds{Left: 0, Top: 0, Width: 1280, Height: 720}
}
