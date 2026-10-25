// Package server exposes the engine over HTTP as described in api/openapi.yaml.
package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/miniaev/hd-calc/internal/ephemeris"
	"github.com/miniaev/hd-calc/internal/hd"
)

type Config struct {
	Engine           *hd.Engine
	EphemerisName    string
	EphemerisVersion string
	SourceURL        string // public repo at the deployed tag (AGPL §13)
}

func New(cfg Config) http.Handler {
	s := &server{cfg: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /v1/version", s.version)
	mux.HandleFunc("POST /v1/chart", s.chart)
	mux.HandleFunc("POST /v1/chart/range", notImplemented)
	mux.HandleFunc("GET /v1/transits", notImplemented)
	mux.HandleFunc("POST /v1/transits/overlay", notImplemented)
	return mux
}

type server struct{ cfg Config }

func (s *server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"engine_version": s.cfg.Engine.EngineVersion,
		"ephemeris": map[string]string{
			"name":    s.cfg.EphemerisName,
			"version": s.cfg.EphemerisVersion,
		},
		"tzdb_version": s.cfg.Engine.TZDBVersion,
		"source_url":   s.cfg.SourceURL,
	})
}

func (s *server) chart(w http.ResponseWriter, r *http.Request) {
	var req hd.ChartRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	c, err := s.cfg.Engine.Chart(req)
	if err != nil {
		writeEngineError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func writeEngineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, hd.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, hd.ErrUnknownTimezone):
		writeError(w, http.StatusUnprocessableEntity, "unknown_timezone", err.Error())
	case errors.Is(err, ephemeris.ErrOutOfRange):
		writeError(w, http.StatusUnprocessableEntity, "out_of_ephemeris_range", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
	}
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotImplemented, "internal", "not implemented yet")
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"code": code, "message": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
