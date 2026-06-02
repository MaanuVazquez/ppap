//go:build windows

package platform

import (
	"fmt"
	"syscall"
)

const dpiAwarenessContextPerMonitorAwareV2 = ^uintptr(3)

var (
	user32                        = syscall.NewLazyDLL("user32.dll")
	procSetProcessDpiAwarenessCtx = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAware        = user32.NewProc("SetProcessDPIAware")
)

func EnableDPIAwareness() error {
	if err := user32.Load(); err != nil {
		return fmt.Errorf("load user32.dll: %w", err)
	}

	if err := procSetProcessDpiAwarenessCtx.Find(); err == nil {
		result, _, callErr := procSetProcessDpiAwarenessCtx.Call(dpiAwarenessContextPerMonitorAwareV2)
		if result != 0 {
			return nil
		}
		return fmt.Errorf("SetProcessDpiAwarenessContext failed: %w", normalizeSyscallError(callErr))
	}

	if err := procSetProcessDPIAware.Find(); err != nil {
		return fmt.Errorf("DPI awareness API unavailable: %w", err)
	}

	result, _, callErr := procSetProcessDPIAware.Call()
	if result == 0 {
		return fmt.Errorf("SetProcessDPIAware failed: %w", normalizeSyscallError(callErr))
	}

	return nil
}

func normalizeSyscallError(err error) error {
	if errno, ok := err.(syscall.Errno); ok && errno == 0 {
		return syscall.EINVAL
	}

	return err
}
