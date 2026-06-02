//go:build windows

package screen

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"

	"ppap/server/internal/display"
)

const (
	biRGB        = 0
	dibRGBColors = 0
	srcCopy      = 0x00CC0020
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	procGetDC                  = user32.NewProc("GetDC")
	procReleaseDC              = user32.NewProc("ReleaseDC")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
)

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [3]uint32
}

type windowsCapturer struct{}

func NewCapturer() (Capturer, error) {
	return &windowsCapturer{}, nil
}

func (capturer *windowsCapturer) Capture() (image.Image, error) {
	bounds := display.VirtualBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return nil, fmt.Errorf("could not read virtual desktop bounds")
	}

	desktopDC, _, err := procGetDC.Call(0)
	if desktopDC == 0 {
		return nil, fmt.Errorf("GetDC failed: %w", err)
	}
	defer procReleaseDC.Call(0, desktopDC)

	memoryDC, _, err := procCreateCompatibleDC.Call(desktopDC)
	if memoryDC == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed: %w", err)
	}
	defer procDeleteDC.Call(memoryDC)

	bitmap, _, err := procCreateCompatibleBitmap.Call(desktopDC, uintptr(bounds.Width), uintptr(bounds.Height))
	if bitmap == 0 {
		return nil, fmt.Errorf("CreateCompatibleBitmap failed: %w", err)
	}
	defer procDeleteObject.Call(bitmap)

	oldObject, _, _ := procSelectObject.Call(memoryDC, bitmap)
	if oldObject != 0 {
		defer procSelectObject.Call(memoryDC, oldObject)
	}

	bltResult, _, err := procBitBlt.Call(
		memoryDC,
		0,
		0,
		uintptr(bounds.Width),
		uintptr(bounds.Height),
		desktopDC,
		intToUintptr(bounds.Left),
		intToUintptr(bounds.Top),
		uintptr(srcCopy),
	)
	if bltResult == 0 {
		return nil, fmt.Errorf("BitBlt failed: %w", err)
	}

	pixelBytes := bounds.Width * bounds.Height * 4
	pixels := make([]byte, pixelBytes)
	info := bitmapInfo{
		Header: bitmapInfoHeader{
			Size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			Width:       int32(bounds.Width),
			Height:      -int32(bounds.Height),
			Planes:      1,
			BitCount:    32,
			Compression: biRGB,
			SizeImage:   uint32(pixelBytes),
		},
	}

	lines, _, err := procGetDIBits.Call(
		memoryDC,
		bitmap,
		0,
		uintptr(bounds.Height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&info)),
		uintptr(dibRGBColors),
	)
	if lines == 0 {
		return nil, fmt.Errorf("GetDIBits failed: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, bounds.Width, bounds.Height))
	for i := 0; i < bounds.Width*bounds.Height; i++ {
		src := i * 4
		dst := i * 4
		img.Pix[dst] = pixels[src+2]
		img.Pix[dst+1] = pixels[src+1]
		img.Pix[dst+2] = pixels[src]
		img.Pix[dst+3] = 255
	}

	return img, nil
}

func (capturer *windowsCapturer) Close() error {
	return nil
}

func intToUintptr(value int) uintptr {
	return uintptr(int32(value))
}
