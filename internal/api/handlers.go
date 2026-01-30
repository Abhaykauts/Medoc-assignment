package api

import (
	"encoding/json"
	"medoc-assignment/internal/core"
	"net/http"
)

type Handler struct {
	Engine core.AllocationEngine
}

func NewHandler(engine core.AllocationEngine) *Handler {
	return &Handler{Engine: engine}
}

// BookRequest payload.
type BookRequest struct {
	DoctorID    string           `json:"doctor_id"`
	SlotID      string           `json:"slot_id"`
	PatientName string           `json:"patient_name"`
	PatientType core.PatientType `json:"patient_type"`
}

func (h *Handler) BookToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.Engine.BookToken(req.DoctorID, req.SlotID, req.PatientName, req.PatientType)
	if err != nil {
		switch err {
		case core.ErrSlotFull:
			http.Error(w, err.Error(), http.StatusConflict)
		case core.ErrDoctorNotFound, core.ErrSlotNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

// GetScheduleHandler returns the current schedule.
func (h *Handler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	doctorID := r.URL.Query().Get("doctor_id")
	if doctorID == "" {
		http.Error(w, "doctor_id is required", http.StatusBadRequest)
		return
	}

	doc, err := h.Engine.GetDoctorSchedule(doctorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// CancelRequest payload.
type CancelRequest struct {
	TokenID string `json:"token_id"`
}

// CancelTokenHandler cancels a token.
func (h *Handler) CancelToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TokenID == "" {
		http.Error(w, "token_id is required", http.StatusBadRequest)
		return
	}

	if err := h.Engine.CancelToken(req.TokenID); err != nil {
		switch err {
		case core.ErrTokenNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"CANCELLED"}`))
}
