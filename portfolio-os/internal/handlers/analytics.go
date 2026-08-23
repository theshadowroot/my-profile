package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"portfolio-os/internal/models"
	"portfolio-os/internal/services"
)

type AnalyticsHandler struct {
	Service *services.AnalyticsService
}

func NewAnalyticsHandler(
	service *services.AnalyticsService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		Service: service,
	}
}

type updateDurationRequest struct {
	DurationSeconds int64 `json:"duration_seconds"`
}

func (h *AnalyticsHandler) StartVisit(
	w http.ResponseWriter,
	r *http.Request,
) {
	visitID, err := services.GenerateVisitID()
	if err != nil {
		http.Error(
			w,
			"failed to create visit",
			http.StatusInternalServerError,
		)
		return
	}

	visit := &models.Visit{
		ID:             visitID,
		IPAddress:      services.AnonymizeIP(services.GetClientIP(r)),
		UserAgent:      r.UserAgent(),
		DeviceCategory: services.DetectDevice(r.UserAgent()),
		Page:           r.URL.Path,
		Timestamp:      time.Now().UTC(),
		Duration:       0,
	}

	h.Service.StartVisit(visit)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{
		"id": visitID,
	}); err != nil {
		return
	}
}

func (h *AnalyticsHandler) UpdateDuration(
	w http.ResponseWriter,
	r *http.Request,
) {
	visitID := r.PathValue("id")

	if visitID == "" {
		http.Error(w, "missing visit ID", http.StatusBadRequest)
		return
	}

	var request updateDurationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if request.DurationSeconds < 0 {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}

	updated := h.Service.UpdateDuration(
		visitID,
		time.Duration(request.DurationSeconds)*time.Second,
	)

	if !updated {
		http.Error(w, "visit not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AnalyticsHandler) GetStats(
	w http.ResponseWriter,
	r *http.Request,
) {
	stats := h.Service.GetStats()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(
			w,
			"failed to encode analytics",
			http.StatusInternalServerError,
		)
	}
}
