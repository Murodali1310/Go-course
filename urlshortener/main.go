//go:build !solution

package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
)

type Server struct {
	mu    sync.Mutex
	store map[string]string
}

func NewServer() *Server {
	return &Server{
		store: make(map[string]string),
	}
}

func main() {
	port := flag.String("port", "8080", "port to serve on")
	flag.Parse()
	fmt.Printf("Serving on port %s\n", *port)

	server := NewServer()
	router := http.NewServeMux()
	router.HandleFunc("/shorten", server.shortenUrl)
	router.HandleFunc("/go/", server.redirectURL)

	http.ListenAndServe(":"+*port, router)
}

type shortenRequest struct {
	Url string `json:"url"`
}

func (s *Server) shortenUrl(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%x", md5.Sum([]byte(req.Url)))
	s.mu.Lock()
	s.store[key] = req.Url
	s.mu.Unlock()

	response := map[string]string{
		"url": req.Url,
		"key": key,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) redirectURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/go/"):]
	s.mu.Lock()
	url, found := s.store[key]
	s.mu.Unlock()
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		http.Redirect(w, r, url, http.StatusFound)
	}
}
