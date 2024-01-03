package web

import (
	"fmt"
	"net/http"
)

func (s *Server) snapshot(w http.ResponseWriter, r *http.Request) {
	if s.snap == nil {
		return
	}
	err := s.snap.Snapshot()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't snapshot: %v", err)
		return
	}
}
