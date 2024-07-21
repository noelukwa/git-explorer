package messaging

import (
	"github.com/noelukwa/git-explorer/internal/events"
)

type Event struct {
	Subject events.EventKind
	Data    []byte
}
