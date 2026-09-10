package handler

import (
	"errors"

	"github.com/m1k1o/neko/server/internal/control"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

var (
	ErrIsNotAllowedToHost = errors.New("is not allowed to host")
	ErrIsNotTheHost       = errors.New("is not the host")
	ErrIsAlreadyTheHost   = errors.New("is already the host")
	ErrIsAlreadyHosted    = errors.New("is already hosted")
)

func (h *MessageHandlerCtx) controlRelease(session types.Session) error {
	err := h.control.Release(session)
	if errors.Is(err, control.ErrNotAllowed) {
		return ErrIsNotAllowedToHost
	}
	if errors.Is(err, control.ErrNotHost) {
		return ErrIsNotTheHost
	}
	return err
}

func (h *MessageHandlerCtx) controlRequest(session types.Session) error {
	result, err := h.control.Request(session)
	if errors.Is(err, control.ErrNotAllowed) {
		return ErrIsNotAllowedToHost
	}
	if errors.Is(err, control.ErrAlreadyHost) {
		return ErrIsAlreadyTheHost
	}
	if err != nil || result.Queued {
		if err != nil {
			return err
		}
		return ErrIsAlreadyHosted
	}
	return nil
}

func (h *MessageHandlerCtx) controlRenew(session types.Session, payload *message.ControlEpoch) error {
	err := h.control.Renew(session, payload.Epoch)
	if errors.Is(err, control.ErrNotAllowed) {
		return ErrIsNotAllowedToHost
	}
	if errors.Is(err, control.ErrNotHost) {
		return ErrIsNotTheHost
	}
	return err
}
