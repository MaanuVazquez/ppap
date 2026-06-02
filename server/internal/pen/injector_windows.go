//go:build windows && amd64

package pen

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"unsafe"

	"ppap/server/internal/display"
	"ppap/server/internal/input"
)

const (
	pointerInputTypePen = 3

	pointerFeedbackNone = 3

	pointerFlagNew        = 0x00000001
	pointerFlagInRange    = 0x00000002
	pointerFlagInContact  = 0x00000004
	pointerFlagFirstBtn   = 0x00000010
	pointerFlagPrimary    = 0x00002000
	pointerFlagConfidence = 0x00004000
	pointerFlagDown       = 0x00010000
	pointerFlagUpdate     = 0x00020000
	pointerFlagUp         = 0x00040000

	pointerChangeNone            = 0
	pointerChangeFirstButtonDown = 1
	pointerChangeFirstButtonUp   = 2

	penMaskPressure = 0x00000001
	penMaskRotation = 0x00000002
	penMaskTiltX    = 0x00000004
	penMaskTiltY    = 0x00000008
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	procCreateSyntheticPointerDevice  = user32.NewProc("CreateSyntheticPointerDevice")
	procInjectSyntheticPointerInput   = user32.NewProc("InjectSyntheticPointerInput")
	procDestroySyntheticPointerDevice = user32.NewProc("DestroySyntheticPointerDevice")
)

type point struct {
	X int32
	Y int32
}

type pointerInfo struct {
	PointerType           uint32
	PointerID             uint32
	FrameID               uint32
	PointerFlags          uint32
	SourceDevice          uintptr
	HwndTarget            uintptr
	PtPixelLocation       point
	PtHimetricLocation    point
	PtPixelLocationRaw    point
	PtHimetricLocationRaw point
	DwTime                uint32
	HistoryCount          uint32
	InputData             int32
	DwKeyStates           uint32
	PerformanceCount      uint64
	ButtonChangeType      int32
}

type pointerPenInfo struct {
	PointerInfo pointerInfo
	PenFlags    uint32
	PenMask     uint32
	Pressure    uint32
	Rotation    uint32
	TiltX       int32
	TiltY       int32
}

type pointerTypeInfo struct {
	Type uint32
	_    uint32
	Pen  pointerPenInfo
}

type windowsInkInjector struct {
	device uintptr
	mutex  sync.Mutex
	isDown bool
}

func NewInjector() (Controller, error) {
	windowsInk, err := newWindowsInkInjector()
	if err != nil {
		return nil, err
	}

	return newController(BackendWindowsInk, []backendEntry{
		{
			info: BackendInfo{
				ID:        BackendWindowsInk,
				Label:     "Windows Ink",
				Available: true,
			},
			injector: windowsInk,
		},
		unavailableBackend(
			BackendWinTab,
			"WinTab",
			"WinTab does not expose a global user-mode injection API; support requires a virtual tablet driver or Wintab32 proxy inside the target app.",
		),
	}), nil
}

func newWindowsInkInjector() (Injector, error) {
	if err := user32.Load(); err != nil {
		return nil, fmt.Errorf("load user32.dll: %w", err)
	}

	device, _, err := procCreateSyntheticPointerDevice.Call(
		uintptr(pointerInputTypePen),
		uintptr(1),
		uintptr(pointerFeedbackNone),
	)
	if device == 0 {
		return nil, fmt.Errorf("CreateSyntheticPointerDevice failed: %w", normalizeSyscallError(err))
	}

	return &windowsInkInjector{device: device}, nil
}

func (injector *windowsInkInjector) Inject(event input.PenEvent) error {
	injector.mutex.Lock()
	defer injector.mutex.Unlock()

	if err := event.Validate(); err != nil {
		return err
	}

	bounds := display.VirtualBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return fmt.Errorf("could not read virtual screen bounds")
	}

	x := int32(bounds.Left + int(math.Round(event.X*float64(bounds.Width-1))))
	y := int32(bounds.Top + int(math.Round(event.Y*float64(bounds.Height-1))))
	pressure := uint32(math.Round(event.Pressure * 1024))

	flags := uint32(pointerFlagInRange | pointerFlagPrimary | pointerFlagConfidence)
	buttonChangeType := int32(pointerChangeNone)
	switch event.Type {
	case input.PenEventDown:
		flags |= pointerFlagNew | pointerFlagInContact | pointerFlagFirstBtn | pointerFlagDown
		pressure = ensureContactPressure(pressure)
		buttonChangeType = pointerChangeFirstButtonDown
		injector.isDown = true
	case input.PenEventMove:
		flags |= pointerFlagUpdate
		if injector.isDown {
			flags |= pointerFlagInContact | pointerFlagFirstBtn
			pressure = ensureContactPressure(pressure)
		} else {
			pressure = 0
		}
	case input.PenEventUp:
		flags |= pointerFlagUp
		pressure = 0
		buttonChangeType = pointerChangeFirstButtonUp
		injector.isDown = false
	}

	location := point{X: x, Y: y}
	penMask := uint32(penMaskPressure)
	rotation := uint32(clampInt(math.Round(event.Twist), 0, 359))
	tiltX := int32(clampInt(math.Round(event.TiltX), -90, 90))
	tiltY := int32(clampInt(math.Round(event.TiltY), -90, 90))
	if rotation != 0 {
		penMask |= penMaskRotation
	}
	if tiltX != 0 {
		penMask |= penMaskTiltX
	}
	if tiltY != 0 {
		penMask |= penMaskTiltY
	}

	packet := pointerTypeInfo{
		Type: pointerInputTypePen,
		Pen: pointerPenInfo{
			PointerInfo: pointerInfo{
				PointerType:        pointerInputTypePen,
				PointerID:          1,
				PointerFlags:       flags,
				PtPixelLocation:    location,
				PtPixelLocationRaw: location,
				HistoryCount:       1,
				ButtonChangeType:   buttonChangeType,
			},
			PenMask:  penMask,
			Pressure: pressure,
			Rotation: rotation,
			TiltX:    tiltX,
			TiltY:    tiltY,
		},
	}

	result, _, err := procInjectSyntheticPointerInput.Call(
		injector.device,
		uintptr(unsafe.Pointer(&packet)),
		uintptr(1),
	)
	if result == 0 {
		return fmt.Errorf("InjectSyntheticPointerInput failed: %w", normalizeSyscallError(err))
	}

	return nil
}

func (injector *windowsInkInjector) Close() error {
	if injector.device != 0 {
		procDestroySyntheticPointerDevice.Call(injector.device)
		injector.device = 0
	}

	return nil
}

func ensureContactPressure(pressure uint32) uint32 {
	if pressure == 0 {
		return 1
	}
	if pressure > 1024 {
		return 1024
	}

	return pressure
}

func clampInt(value float64, min int, max int) int {
	if value < float64(min) {
		return min
	}
	if value > float64(max) {
		return max
	}

	return int(value)
}

func normalizeSyscallError(err error) error {
	if errno, ok := err.(syscall.Errno); ok && errno == 0 {
		return syscall.EINVAL
	}

	return err
}
