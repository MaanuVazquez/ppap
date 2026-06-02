//go:build !windows

package platform

func ActiveWindowInfo() WindowInfo {
	return WindowInfo{}
}
