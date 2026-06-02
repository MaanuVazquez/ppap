package pen

import "ppap/server/internal/input"

type Injector interface {
	Inject(event input.PenEvent) error
	Close() error
}
