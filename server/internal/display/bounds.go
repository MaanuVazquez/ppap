package display

type Bounds struct {
	Left   int
	Top    int
	Width  int
	Height int
}

func (bounds Bounds) Right() int {
	return bounds.Left + bounds.Width
}

func (bounds Bounds) Bottom() int {
	return bounds.Top + bounds.Height
}
