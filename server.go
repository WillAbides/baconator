package baconator

import (
	"encoding/json"
	"net/http"
)

// Server is an http server for baconator
type Server struct {
	baconator *Baconator
}

// NewServer returns a new Server
func NewServer(baconator *Baconator) *Server {
	return &Server{
		baconator: baconator,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	switch req.URL.Path {
	case "/center":
		s.center(w, req)
	case "/link":
		s.link(w, req)
	default:
		http.Error(w, "", http.StatusNotFound)
	}
}

func (s *Server) link(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	src := query.Get("a")
	if src == "" {
		http.Error(w, "a is a required query parameter", http.StatusBadRequest)
		return
	}
	dest := query.Get("b")
	if dest == "" {
		http.Error(w, "b is a required query parameter", http.StatusBadRequest)
		return
	}
	res, err := s.baconator.links(src, dest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		panic(err)
	}
}

func (s *Server) center(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Query().Get("p")
	if p == "" {
		http.Error(w, "p is a required query parameter", http.StatusBadRequest)
		return
	}
	res := s.baconator.center(s.baconator.CastNodes[p])
	if res == nil {
		http.Error(w, "person not found", http.StatusNotFound)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		panic(err)
	}
}
