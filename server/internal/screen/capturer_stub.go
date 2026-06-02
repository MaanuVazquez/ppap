//go:build !windows

package screen

import (
	"image"
	"image/color"
)

type stubCapturer struct{}

func NewCapturer() (Capturer, error) {
	return &stubCapturer{}, nil
}

func (capturer *stubCapturer) Capture() (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	bg := color.RGBA{R: 13, G: 17, B: 23, A: 255}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.SetRGBA(x, y, bg)
		}
	}

	return img, nil
}

func (capturer *stubCapturer) Close() error {
	return nil
}
