//go:build windows && amd64

package pen

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"unsafe"

	"ppap/server/internal/input"
)

const (
	ppapWinTabMappingName = "Local\\PPAPWinTabPacketV1"
	ppapWinTabMagic       = 0x50504157
	ppapWinTabVersion     = 1

	pageReadWrite    = 0x04
	fileMapAllAccess = 0x000F001F
)

var (
	winTabKernel32         = syscall.NewLazyDLL("kernel32.dll")
	procCreateFileMappingW = winTabKernel32.NewProc("CreateFileMappingW")
	procMapViewOfFile      = winTabKernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile    = winTabKernel32.NewProc("UnmapViewOfFile")
	procCloseHandle        = winTabKernel32.NewProc("CloseHandle")
)

type winTabSharedPacket struct {
	Magic     uint32
	Version   uint32
	Serial    uint32
	X         int32
	Y         int32
	Buttons   uint32
	Pressure  uint32
	InContact uint32
}

type winTabBridgeInjector struct {
	mutex   sync.Mutex
	mapping uintptr
	packet  *winTabSharedPacket
	serial  uint32
}

func newWinTabBridgeInjector() (Injector, error) {
	mappingName, err := syscall.UTF16PtrFromString(ppapWinTabMappingName)
	if err != nil {
		return nil, err
	}

	mapping, _, callErr := procCreateFileMappingW.Call(
		^uintptr(0),
		0,
		uintptr(pageReadWrite),
		0,
		unsafe.Sizeof(winTabSharedPacket{}),
		uintptr(unsafe.Pointer(mappingName)),
	)
	if mapping == 0 {
		return nil, fmt.Errorf("CreateFileMappingW failed: %w", normalizeSyscallError(callErr))
	}

	view, _, callErr := procMapViewOfFile.Call(
		mapping,
		uintptr(fileMapAllAccess),
		0,
		0,
		unsafe.Sizeof(winTabSharedPacket{}),
	)
	if view == 0 {
		procCloseHandle.Call(mapping)
		return nil, fmt.Errorf("MapViewOfFile failed: %w", normalizeSyscallError(callErr))
	}

	injector := &winTabBridgeInjector{
		mapping: mapping,
		packet:  (*winTabSharedPacket)(unsafe.Pointer(view)),
	}
	injector.packet.Magic = ppapWinTabMagic
	injector.packet.Version = ppapWinTabVersion

	return injector, nil
}

func (injector *winTabBridgeInjector) Inject(event input.PenEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	injector.mutex.Lock()
	defer injector.mutex.Unlock()

	injector.serial++
	injector.packet.Magic = ppapWinTabMagic
	injector.packet.Version = ppapWinTabVersion
	injector.packet.X = int32(math.Round(event.X * 65535))
	injector.packet.Y = int32(math.Round(event.Y * 65535))
	injector.packet.Pressure = uint32(math.Round(event.Pressure * 1024))

	switch event.Type {
	case input.PenEventDown, input.PenEventMove:
		injector.packet.InContact = 1
		injector.packet.Buttons = 1
	case input.PenEventUp:
		injector.packet.InContact = 0
		injector.packet.Buttons = 0
		injector.packet.Pressure = 0
	}

	// Serial is written last so proxy readers see a complete packet per serial.
	injector.packet.Serial = injector.serial
	return nil
}

func (injector *winTabBridgeInjector) Close() error {
	if injector.packet != nil {
		procUnmapViewOfFile.Call(uintptr(unsafe.Pointer(injector.packet)))
		injector.packet = nil
	}
	if injector.mapping != 0 {
		procCloseHandle.Call(injector.mapping)
		injector.mapping = 0
	}

	return nil
}
