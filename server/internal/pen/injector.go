package pen

import (
	"errors"
	"fmt"
	"sync"

	"ppap/server/internal/input"
)

var errWinTabBridgeUnavailable = errors.New("WinTab bridge requires windows/amd64")

type Backend string

const (
	BackendWindowsInk Backend = "windowsInk"
	BackendWinTab     Backend = "winTab"
)

type BackendInfo struct {
	ID        Backend `json:"id"`
	Label     string  `json:"label"`
	Available bool    `json:"available"`
	Reason    string  `json:"reason,omitempty"`
}

type Injector interface {
	Inject(event input.PenEvent) error
	Close() error
}

type Controller interface {
	Injector
	ActiveBackend() Backend
	Backends() []BackendInfo
	SetBackend(backend Backend) error
}

type backendEntry struct {
	info     BackendInfo
	injector Injector
}

type controller struct {
	mutex    sync.RWMutex
	active   Backend
	backends map[Backend]backendEntry
}

func newController(active Backend, entries []backendEntry) Controller {
	backends := make(map[Backend]backendEntry, len(entries))
	for _, entry := range entries {
		backends[entry.info.ID] = entry
	}

	return &controller{active: active, backends: backends}
}

func (controller *controller) Inject(event input.PenEvent) error {
	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	entry, err := controller.activeEntry()
	if err != nil {
		return err
	}
	if !entry.info.Available || entry.injector == nil {
		return fmt.Errorf("%s backend unavailable: %s", entry.info.Label, entry.info.Reason)
	}

	return entry.injector.Inject(event)
}

func (controller *controller) Close() error {
	var closeErr error
	for _, entry := range controller.backends {
		if entry.injector == nil {
			continue
		}
		if err := entry.injector.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}

	return closeErr
}

func (controller *controller) ActiveBackend() Backend {
	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	return controller.active
}

func (controller *controller) Backends() []BackendInfo {
	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	infos := make([]BackendInfo, 0, len(controller.backends))
	for _, backend := range []Backend{BackendWindowsInk, BackendWinTab} {
		entry, ok := controller.backends[backend]
		if ok {
			infos = append(infos, entry.info)
		}
	}

	return infos
}

func (controller *controller) SetBackend(backend Backend) error {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	entry, ok := controller.backends[backend]
	if !ok {
		return fmt.Errorf("unknown pen backend %q", backend)
	}
	if !entry.info.Available {
		return fmt.Errorf("%s backend unavailable: %s", entry.info.Label, entry.info.Reason)
	}

	controller.active = backend
	return nil
}

func (controller *controller) activeEntry() (backendEntry, error) {
	entry, ok := controller.backends[controller.active]
	if !ok {
		return backendEntry{}, fmt.Errorf("active pen backend %q is not registered", controller.active)
	}

	return entry, nil
}

func unavailableBackend(id Backend, label string, reason string) backendEntry {
	return backendEntry{
		info: BackendInfo{
			ID:        id,
			Label:     label,
			Available: false,
			Reason:    reason,
		},
	}
}
