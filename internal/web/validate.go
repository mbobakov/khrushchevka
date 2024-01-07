package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mbobakov/khrushchevka/internal"
)

type board struct {
	ID   uint8
	View string
}

type pin struct {
	ID   string
	IsOn bool
}

type validateContext struct {
	Active      string
	Boards      []board
	ActiveBoard uint8
	APins       []pin
	BPins       []pin
}

func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
	// set flow to manual
	err := s.flows.SelectFlow(s.mainCtx, "manual")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't select manual flow for validation: %v", err)
		return
	}

	vctx, err := s.validateContext(s.lights.Boards(), 0)
	if err != nil {
		fmt.Fprintf(w, "couldn't build index context: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}

	err = s.indexTmpl.ExecuteTemplate(buf, "validate.gotmpl", vctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't execute template: %v", err)
		return
	}

	w.Write(buf.Bytes()) //nolint: errcheck
}

func (s *Server) validatePost(w http.ResponseWriter, r *http.Request) {

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "couldn't read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params, err := url.ParseQuery(string(buf))
	if err != nil {
		fmt.Fprintf(w, "couldn't read params: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	boardRaw := params.Get("board")
	board, err := strconv.Atoi(boardRaw)
	if err != nil {
		fmt.Fprintf(w, "couldn't read board parameter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pinsToOn := params["pin"]
	if s.validateSelectBoard == 0 || s.validateSelectBoard == uint8(board) {
		for _, p := range pinsToOn {
			err = s.lights.Set(internal.LightAddress{
				Pin:   p,
				Board: uint8(board),
			}, true)

			if err != nil {
				fmt.Fprintf(w, "couldn't set pin %s on board 0x%x: %v", p, board, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	s.validateSelectBoard = uint8(board)

	vctx, err := s.validateContext(s.lights.Boards(), uint8(board))
	if err != nil {
		fmt.Fprintf(w, "couldn't build index context: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bufResp := &bytes.Buffer{}

	err = s.indexTmpl.ExecuteTemplate(bufResp, "validate-form.gotmpl", vctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't execute template: %v", err)
		return
	}

	w.Write(bufResp.Bytes()) //nolint: errcheck
}

func (s *Server) validateContext(boards []uint8, active uint8) (*validateContext, error) {
	apins := []string{"A7", "A6", "A5", "A4", "A3", "A2", "A1", "A0"}
	bpins := []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"}

	result := &validateContext{
		Active:      "validate",
		ActiveBoard: active,
		APins:       []pin{},
		BPins:       []pin{},
	}
	if active != 0 {
		for _, p := range apins {
			isOn, err := s.lights.IsOn(internal.LightAddress{
				Pin:   p,
				Board: active,
			})
			if err != nil {
				return nil, fmt.Errorf("couldn't get pin %s on board 0x%x: %v", p, active, err)
			}
			result.APins = append(result.APins, pin{
				ID:   p,
				IsOn: isOn,
			})
		}
		for _, p := range bpins {
			isOn, err := s.lights.IsOn(internal.LightAddress{
				Pin:   p,
				Board: active,
			})
			if err != nil {
				return nil, fmt.Errorf("couldn't get pin %s on board 0x%x: %v", p, active, err)
			}
			result.BPins = append(result.BPins, pin{
				ID:   p,
				IsOn: isOn,
			})
		}
	}

	for _, b := range boards {
		result.Boards = append(result.Boards, board{
			ID:   b,
			View: fmt.Sprintf("0x%x", b),
		})
	}

	return result, nil
}
