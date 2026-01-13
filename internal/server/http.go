package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nuggetplum/MiniKafka/internal/log" // Import your store package
)

type httpServer struct {
	Log *log.Store
}

// newHTTPServer is a factory that creates the listener
func newHTTPServer(log *log.Store) *httpServer {
	return &httpServer{
		Log: log,
	}
}

// Request/Response structs to keep our API clean
type ProduceRequest struct {
	Record log.Record `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record log.Record `json:"record"`
}

func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the request
	var req ProduceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Call the core logic (The Store)
	off, err := s.Log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Encode the response
	res := ProduceResponse{Offset: off}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	// var req ConsumeRequest
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		http.Error(w, "offset is required", http.StatusBadRequest)
		return
	}

	offset, err := strconv.ParseUint(offsetStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid offset format", http.StatusBadRequest)
		return
	}

	record, err := s.Log.Read(offset)
	if err == log.ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusNotFound) // 404 if offset doesn't exist
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConsumeResponse{Record: record}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewHTTPServer creates the server and sets up routes
func NewHTTPServer(addr string, log *log.Store) *http.Server {
	httpsrv := newHTTPServer(log)

	mux := http.NewServeMux()
	// Map endpoints to functions
	mux.HandleFunc("POST /", httpsrv.handleProduce)
	mux.HandleFunc("GET /", httpsrv.handleConsume)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
