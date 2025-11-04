package handlers

import (
	"encoding/json"
	"github.com/DrummDaddy/Booking_service/internal/models"
	"github.com/DrummDaddy/Booking_service/internal/services"
	"net/http"
)

type BookingHandler struct {
	Service *services.BookingService
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Incorrect request", http.StatusBadRequest)
		return
	}
	userID := r.Header.Get("X-USER-ID")
	req.UserID = userID

	resp, err := h.Service.CreatingBooking(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (h *BookingHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	type paymentReq struct {
		BookingID string `json:"booking_id"`
		ReturnURL string `json:"return_url"`
	}
	var req paymentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Incorrect request", http.StatusBadRequest)
		return
	}

	paymentURL, err := h.Service.CreatePayment(r.Context(), req.BookingID, req.ReturnURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(paymentURL)
}
