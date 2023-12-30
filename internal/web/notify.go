package web

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/mbobakov/khrushchevka/internal"
	"github.com/r3labs/sse"
)

func (s *Server) NotifyViaSSE(ctx context.Context) error {
	ch := make(chan internal.PinState)
	s.lights.Subscribe(ch)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping SSE notifications")
			return nil
		case pin := <-ch:
			slog.Debug("got pin state", slog.Any("pin", pin))
			lctx, err := s.lightContextByPinState(pin)
			if err != nil {
				slog.Error("couldn't get light context", slog.Any("err", err))
				continue
			}
			buf := &bytes.Buffer{}

			err = s.indexTmpl.ExecuteTemplate(buf, "light.gotmpl", lctx)

			if err != nil {
				slog.Error("couldn't execute template", slog.Any("err", err))
				continue
			}

			s.sse.Publish("lights", &sse.Event{
				Event: []byte(lctx.ID),
				Data:  bytes.ReplaceAll(buf.Bytes(), []byte("\n"), []byte("")),
			})
		}
	}
}

func (s *Server) lightContextByPinState(pin internal.PinState) (*lightContext, error) {
	for _, ll := range s.mapping {
		for _, wnd := range ll {
			if wnd.Addr.Board != pin.Addr.Board || wnd.Addr.Pin != pin.Addr.Pin {
				continue
			}

			return &lightContext{
				ID:         lightID(wnd),
				IsOn:       pin.IsOn,
				FlatNumber: wnd.Number,
				Class:      cssClassByType(wnd.Kind),
				Addr:       wnd.Addr,
			}, nil
		}
	}

	return nil, fmt.Errorf("couldn't find light with board '%d' pin %s", pin.Addr.Board, pin.Addr.Pin)
}
