package platform

type WindowInfo struct {
	Handle uintptr `json:"handle"`
	Title  string  `json:"title"`
	Class  string  `json:"class"`
}
