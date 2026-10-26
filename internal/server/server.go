// Package server exposes the engine over HTTP as described in api/openapi.yaml.
package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
	mux.HandleFunc("POST /v1/chart/range", s.chartRange)
	mux.HandleFunc("GET /v1/transits", s.transits)
	mux.HandleFunc("POST /v1/transits/overlay", s.overlay)
	mux.HandleFunc("POST /v1/composite", s.composite)
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

func (s *server) chartRange(w http.ResponseWriter, r *http.Request) {
	var req hd.ChartRangeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	c, err := s.cfg.Engine.ChartRange(req)
	if err != nil {
		writeEngineError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *server) transits(w http.ResponseWriter, r *http.Request) {
	at, err := time.Parse(time.RFC3339, r.URL.Query().Get("at"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "at: want RFC 3339 date-time")
		return
	}
	t, err := s.cfg.Engine.Transits(at)
	if err != nil {
		writeEngineError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *server) overlay(w http.ResponseWriter, r *http.Request) {
	var req hd.TransitOverlayRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	o, err := s.cfg.Engine.Overlay(req)
	if err != nil {
		writeEngineError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *server) composite(w http.ResponseWriter, r *http.Request) {
	var req hd.CompositeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	c, err := s.cfg.Engine.Composite(req)
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
	case errors.Is(err, hd.ErrIntervalTooLong):
		writeError(w, http.StatusUnprocessableEntity, "interval_too_long", err.Error())
	case errors.Is(err, hd.ErrUnknownTimezone):
		writeError(w, http.StatusUnprocessableEntity, "unknown_timezone", err.Error())
	case errors.Is(err, ephemeris.ErrOutOfRange):
		writeError(w, http.StatusUnprocessableEntity, "out_of_ephemeris_range", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"code": code, "message": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
