package screen

import "image"

type Capturer interface {
	Capture() (image.Image, error)
	Close() error
}
