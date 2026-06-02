//go:build windows

package display

import "syscall"

const (
	smXVirtualScreen  = 76
	smYVirtualScreen  = 77
	smCXVirtualScreen = 78
	smCYVirtualScreen = 79
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
)

func VirtualBounds() Bounds {
	left, _, _ := procGetSystemMetrics.Call(uintptr(smXVirtualScreen))
	top, _, _ := procGetSystemMetrics.Call(uintptr(smYVirtualScreen))
	width, _, _ := procGetSystemMetrics.Call(uintptr(smCXVirtualScreen))
	height, _, _ := procGetSystemMetrics.Call(uintptr(smCYVirtualScreen))

	return Bounds{
		Left:   int(int32(left)),
		Top:    int(int32(top)),
		Width:  int(width),
		Height: int(height),
	}
}
