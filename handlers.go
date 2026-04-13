package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	db *sql.DB
}

func (s *Server) handleGetActivePromos(w http.ResponseWriter, r *http.Request) {
	promos, err := getActivePromos(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(promos)
}

type createPromoRequest struct {
	StationName      string  `json:"station_name"`
	SourceMessageID  *string `json:"source_message_id"`
}

type createPromoResponse struct {
	StationID int64     `json:"station_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) handleCreatePromo(w http.ResponseWriter, r *http.Request) {
	var req createPromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.StationName == "" {
		http.Error(w, "station_name required", http.StatusBadRequest)
		return
	}

	stationID, err := getStationIDByName(s.db, req.StationName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "station not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, createdAt, err := insertPromo(s.db, stationID, req.SourceMessageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createPromoResponse{
		StationID: id,
		CreatedAt: createdAt,
	})
}

func (s *Server) handleUpdatePromo(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/promos/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := updatePromo(s.db, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeletePromo(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/promos/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := deletePromo(s.db, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
