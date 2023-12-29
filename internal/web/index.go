package web

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/mbobakov/khrushchevka/internal"
)

type lightContext struct {
	ID         string
	IsOn       bool
	FlatNumber int
	Class      string
	Addr       internal.LightAddress
}

type indexContext struct {
	Front [][]*lightContext
	Right [][]*lightContext
	Back  [][]*lightContext
	Left  [][]*lightContext
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	ictx, err := s.indexContext(s.mapping)
	if err != nil {
		fmt.Fprintf(w, "couldn't build index context: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}

	err = s.indexTmpl.Execute(buf, ictx)
	if err != nil {
		fmt.Fprintf(w, "couldn't execute template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

func (s *Server) indexContext(mapping [][]internal.Light) (*indexContext, error) {
	result := &indexContext{}

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
		isOn, err = s.lights.IsOn(l.Addr.Board, l.Addr.Pin)
		if err != nil {
			return nil, fmt.Errorf("couldn't get status of the light: %w", err)
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
