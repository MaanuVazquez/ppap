//go:build !windows

package platform

func EnableDPIAwareness() error {
	return nil
}
