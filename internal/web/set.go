package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mbobakov/khrushchevka/internal"
)

func (s *Server) setLigts(w http.ResponseWriter, r *http.Request) {
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

	isOnRaw := params.Get("is_on")
	mustOn, err := strconv.ParseBool(isOnRaw)
	if err != nil {
		fmt.Fprintf(w, "couldn't read isOn parameter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pin := params.Get("pin")

	light := internal.Light{}

SEARCH:
	for _, lvl := range s.mapping {
		for _, l := range lvl {
			if l.Addr.Board == uint8(board) && l.Addr.Pin == pin {
				light = l
				break SEARCH
			}
		}

	}

	if light.Addr.Board == 0 {
		fmt.Fprintf(w, "couldn't find light with board %d and pin %s", board, pin)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.lights.Set(internal.LightAddress{
		Pin:   pin,
		Board: uint8(board),
	}, mustOn)

	if err != nil {
		fmt.Fprintf(w, "couldn't set lights: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lctx, err := s.lightContext(light)
	if err != nil {
		fmt.Fprintf(w, "couldn't build light context: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err != nil {
		fmt.Fprintf(w, "couldn't get window state: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.indexTmpl.ExecuteTemplate(w, "light.gotmpl", lctx)
	if err != nil {
		fmt.Fprintf(w, "couldn't execute template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
