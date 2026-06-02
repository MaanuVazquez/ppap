//go:build windows

package platform

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	procGetForegroundWindow  = user32.NewProc("GetForegroundWindow")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetClassNameW        = user32.NewProc("GetClassNameW")
)

func ActiveWindowInfo() WindowInfo {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return WindowInfo{}
	}

	return WindowInfo{
		Handle: hwnd,
		Title:  windowText(hwnd),
		Class:  windowClass(hwnd),
	}
}

func windowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}

	buffer := make([]uint16, int(length)+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return strings.TrimRight(syscall.UTF16ToString(buffer), "\x00")
}

func windowClass(hwnd uintptr) string {
	buffer := make([]uint16, 256)
	procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return strings.TrimRight(syscall.UTF16ToString(buffer), "\x00")
}
