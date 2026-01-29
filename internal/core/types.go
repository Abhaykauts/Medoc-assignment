package core

import (
	"time"
)

// PatientType represents the priority of the patient.
type PatientType string

const (
	Emergency     PatientType = "EMERGENCY"
	PaidPriority  PatientType = "PAID_PRIORITY"
	FollowUp      PatientType = "FOLLOW_UP"
	OnlineBooking PatientType = "ONLINE_BOOKING" // Also Walk-in
)

// Priority returns an integer representation of the patient type (higher is better).
func (pt PatientType) Priority() int {
	switch pt {
	case Emergency:
		return 100
	case PaidPriority:
		return 80
	case FollowUp:
		return 60
	case OnlineBooking:
		return 40
	default:
		return 0
	}
}

// Token represents an allocated appointment.
type Token struct {
	ID          string      `json:"id"`
	PatientName string      `json:"patient_name"`
	Type        PatientType `json:"type"`
	Timestamp   time.Time   `json:"timestamp"` // Creation time
	SlotID      string      `json:"slot_id"`
	DoctorID    string      `json:"doctor_id"`
	Status      string      `json:"status"` // BOOKED, CANCELLED, COMPLETED
}

// Slot represents a fixed time window for a doctor.
type Slot struct {
	ID        string    `json:"id"`
	DoctorID  string    `json:"doctor_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Capacity  int       `json:"capacity"`
	tokens    []*Token  // Private to ensure encapsulation via methods if needed
}

// NewSlot creates a new slot.
func NewSlot(id, doctorID string, start, end time.Time, capacity int) *Slot {
	return &Slot{
		ID:        id,
		DoctorID:  doctorID,
		StartTime: start,
		EndTime:   end,
		Capacity:  capacity,
		tokens:    make([]*Token, 0, capacity),
	}
}

// Tokens returns a copy of tokens allocated to this slot.
func (s *Slot) Tokens() []*Token {
	// Return copy to prevent external mutation issues
	result := make([]*Token, len(s.tokens))
	copy(result, s.tokens)
	return result
}

// AddToken adds a token to the slot. Return error if full? 
// The AllocationEngine will handle logic, this just holds data.
func (s *Slot) AddToken(t *Token) {
	s.tokens = append(s.tokens, t)
}

// RemoveToken removes a token by ID.
func (s *Slot) RemoveToken(tokenID string) *Token {
	for i, t := range s.tokens {
		if t.ID == tokenID {
			s.tokens = append(s.tokens[:i], s.tokens[i+1:]...)
			return t
		}
	}
	return nil
}

// CurrentCount returns current number of tokens.
func (s *Slot) CurrentCount() int {
	return len(s.tokens)
}

// Doctor represents a medical professional.
type Doctor struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Slots []*Slot `json:"slots"`
}
