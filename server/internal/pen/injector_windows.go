//go:build windows && amd64

package pen

import (
	"fmt"
	"math"
	"syscall"
	"unsafe"

	"ppap/server/internal/input"
)

const (
	pointerInputTypePen = 3

	pointerFeedbackDefault = 1

	pointerFlagInRange   = 0x00000002
	pointerFlagInContact = 0x00000004
	pointerFlagDown      = 0x00010000
	pointerFlagUpdate    = 0x00020000
	pointerFlagUp        = 0x00040000

	penMaskPressure = 0x00000001
	penMaskRotation = 0x00000002
	penMaskTiltX    = 0x00000004
	penMaskTiltY    = 0x00000008

	smCXScreen = 0
	smCYScreen = 1
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	procCreateSyntheticPointerDevice  = user32.NewProc("CreateSyntheticPointerDevice")
	procInjectSyntheticPointerInput   = user32.NewProc("InjectSyntheticPointerInput")
	procDestroySyntheticPointerDevice = user32.NewProc("DestroySyntheticPointerDevice")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
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

type windowsInjector struct {
	device uintptr
}

func NewInjector() (Injector, error) {
	if err := user32.Load(); err != nil {
		return nil, fmt.Errorf("load user32.dll: %w", err)
	}

	device, _, err := procCreateSyntheticPointerDevice.Call(
		uintptr(pointerInputTypePen),
		uintptr(1),
		uintptr(pointerFeedbackDefault),
	)
	if device == 0 {
		return nil, fmt.Errorf("CreateSyntheticPointerDevice failed: %w", err)
	}

	return &windowsInjector{device: device}, nil
}

func (injector *windowsInjector) Inject(event input.PenEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	width, height := screenSize()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("could not read screen size")
	}

	x := int32(math.Round(event.X * float64(width-1)))
	y := int32(math.Round(event.Y * float64(height-1)))
	pressure := uint32(math.Round(event.Pressure * 1024))

	flags := uint32(pointerFlagInRange)
	switch event.Type {
	case input.PenEventDown:
		flags |= pointerFlagInContact | pointerFlagDown
		pressure = ensureContactPressure(pressure)
	case input.PenEventMove:
		flags |= pointerFlagInContact | pointerFlagUpdate
		pressure = ensureContactPressure(pressure)
	case input.PenEventUp:
		flags |= pointerFlagUp
		pressure = 0
	}

	packet := pointerTypeInfo{
		Type: pointerInputTypePen,
		Pen: pointerPenInfo{
			PointerInfo: pointerInfo{
				PointerType:     pointerInputTypePen,
				PointerID:       1,
				PointerFlags:    flags,
				PtPixelLocation: point{X: x, Y: y},
				HistoryCount:    1,
			},
			PenMask:  penMaskPressure | penMaskRotation | penMaskTiltX | penMaskTiltY,
			Pressure: pressure,
			Rotation: uint32(clampInt(math.Round(event.Twist), 0, 359)),
			TiltX:    int32(clampInt(math.Round(event.TiltX), -90, 90)),
			TiltY:    int32(clampInt(math.Round(event.TiltY), -90, 90)),
		},
	}

	result, _, err := procInjectSyntheticPointerInput.Call(
		injector.device,
		uintptr(unsafe.Pointer(&packet)),
		uintptr(1),
	)
	if result == 0 {
		return fmt.Errorf("InjectSyntheticPointerInput failed: %w", err)
	}

	return nil
}

func (injector *windowsInjector) Close() error {
	if injector.device != 0 {
		procDestroySyntheticPointerDevice.Call(injector.device)
		injector.device = 0
	}

	return nil
}

func screenSize() (int, int) {
	width, _, _ := procGetSystemMetrics.Call(uintptr(smCXScreen))
	height, _, _ := procGetSystemMetrics.Call(uintptr(smCYScreen))

	return int(width), int(height)
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
