package web

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/mbobakov/khrushchevka/internal"
)

type lightContext struct {
	ID         string
	IsOn       bool
	FlatNumber int
	Class      string
	Addr       internal.LightAddress
}
type flowContext struct {
	Names    []string
	Selected string
}

type indexContext struct {
	Active string
	Front  [][]*lightContext
	Right  [][]*lightContext
	Back   [][]*lightContext
	Left   [][]*lightContext
	Flows  *flowContext
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	ictx, err := s.indexContext(s.mapping)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't build index context: %v", err)
		return
	}

	buf := &bytes.Buffer{}

	err = s.indexTmpl.ExecuteTemplate(buf, "index.gotmpl", ictx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't execute template: %v", err)
		return
	}

	w.Write(buf.Bytes()) //nolint: errcheck
}

func (s *Server) indexContext(mapping [][]internal.Light) (*indexContext, error) {
	result := &indexContext{
		Active: "index",
		Flows: &flowContext{
			Names:    s.flows.FlowNames(),
			Selected: s.flows.Active(),
		},
	}

	for _, ll := range mapping {
		floorFront := []*lightContext{} // Front side itis where we ends are meet it's why we have to change direction a litte bit
		floorRight := []*lightContext{}
		floorBack := []*lightContext{}
		floorLeft := []*lightContext{}

		for _, wnd := range ll {
			lctx, err := s.lightContext(wnd)
			if err != nil {
				return nil, fmt.Errorf("couldn't build light context: %w", err)
			}
			switch wnd.Side {
			case internal.SideFront:
				floorFront = append(floorFront, lctx)
			case internal.SideRight:
				floorRight = append(floorRight, lctx)
			case internal.SideBack:
				floorBack = append(floorBack, lctx)
			case internal.SideLeft:
				floorLeft = append(floorLeft, lctx)
			}
		}

		result.Front = append(result.Front, floorFront)
		result.Right = append(result.Right, floorRight)
		result.Back = append(result.Back, floorBack)
		result.Left = append(result.Left, floorLeft)
	}

	slices.Reverse(result.Front)
	slices.Reverse(result.Right)
	slices.Reverse(result.Back)
	slices.Reverse(result.Left)

	return result, nil
}

func cssClassByType(t internal.LightType) string {
	switch t {
	case internal.LightTypeServiceNoManLand:
		return "nomanland"
	case internal.LightTypeServiceEntrance:
		return "entrance"
	case internal.LightTypeWallStub:
		return "wall"
	case internal.LightTypeShortWindow:
		return "short"
	case internal.LightTypeLongWindow:
		return "long"
	default:
		return ""
	}
}

func (s *Server) lightContext(l internal.Light) (*lightContext, error) {
	var (
		isOn bool
		err  error
	)

	if l.Addr.Pin != "" {
		isOn, err = s.lights.IsOn(l.Addr)
		if err != nil {
			slog.Error("couldn't get status of the light", slog.String("err", err.Error()))
			return nil, nil
		}
	}
	return &lightContext{
		ID:         lightID(l),
		IsOn:       isOn,
		FlatNumber: l.Number,
		Class:      cssClassByType(l.Kind),
		Addr:       l.Addr,
	}, nil

}
