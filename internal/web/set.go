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

	flatRaw := params.Get("flat_number")
	flat, err := strconv.Atoi(flatRaw)
	if err != nil {
		fmt.Fprintf(w, "couldn't read flat parameter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isOnRaw := params.Get("is_on")
	isOn, err := strconv.ParseBool(isOnRaw)
	if err != nil {
		fmt.Fprintf(w, "couldn't read isOn parameter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pin := params.Get("pin")
	class := params.Get("class")
	id := params.Get("id")

	err = s.lights.Set(internal.LightAddress{
		Pin:   pin,
		Board: uint8(board),
	}, !isOn)

	if err != nil {
		fmt.Fprintf(w, "couldn't set lights: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lctx := &lightContext{
		ID:         id,
		IsOn:       !isOn,
		FlatNumber: flat,
		Class:      class,
		Addr: internal.LightAddress{
			Pin:   pin,
			Board: uint8(board),
		},
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
