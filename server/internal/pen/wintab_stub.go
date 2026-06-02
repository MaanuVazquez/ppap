//go:build !windows || !amd64

package pen

func newWinTabBridgeInjector() (Injector, error) {
	return nil, errWinTabBridgeUnavailable
}
