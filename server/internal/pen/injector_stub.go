//go:build !windows || !amd64

package pen

import "fmt"

func NewInjector() (Injector, error) {
	return nil, fmt.Errorf("synthetic Windows Ink pen injection requires windows/amd64")
}
