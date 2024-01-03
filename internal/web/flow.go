package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (s *Server) setFlow(w http.ResponseWriter, r *http.Request) {
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

	active := params.Get("selected")

	err = s.flows.SelectFlow(s.mainCtx, active)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "couldn't select Flow: %v", err)
		return
	}

	fctx := &flowContext{
		Names:    s.flows.FlowNames(),
		Selected: active,
	}

	err = s.indexTmpl.ExecuteTemplate(w, "flows.gotmpl", fctx)
	if err != nil {
		fmt.Fprintf(w, "couldn't execute template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
